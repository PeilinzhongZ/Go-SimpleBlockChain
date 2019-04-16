package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Flag_value struct {
	Encoded_prefix []uint8
	Value          string
}

type Node struct {
	Node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	Branch_value [17]string
	Flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	Db     map[string]Node
	Root   string
	Values map[string]string
}

// func (mpt *MerklePatriciaTrie) GetAll() map[string]string {
//
// }

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	// TODO
	if mpt.Root != "" {
		keyHex := append(keyToHexArr(key), uint8(16))
		value := mpt.get(mpt.Root, keyHex)
		if value == "" {
			return "", errors.New("path_not_found")
		}
		return value, nil
	}
	return "", errors.New("path_not_found")
}

func (mpt *MerklePatriciaTrie) get(hashValue string, keyHex []uint8) string {
	node := mpt.Db[hashValue]
	if node.Node_type == 2 {
		path := compact_decode(node.Flag_value.Encoded_prefix)
		nodeType := node.Flag_value.Encoded_prefix[0] / uint8(16)
		if nodeType == uint8(0) || nodeType == uint8(1) {
			//Ext
			prefixIndex := findPrefixIndex(keyHex, path)
			if prefixIndex != len(path)-1 {
				return ""
			}
			return mpt.get(node.Flag_value.Value, keyHex[prefixIndex+1:])
		}
		//leaf
		path = append(path, uint8(16))
		prefixIndex := findPrefixIndex(keyHex, path)
		if prefixIndex == len(path)-1 && prefixIndex == len(keyHex)-1 {
			return node.Flag_value.Value
		}
		return ""
	}
	//branch
	if node.Branch_value[keyHex[0]] == "" {
		return ""
	} else if len(keyHex) == 1 {
		return node.Branch_value[16]
	}
	return mpt.get(node.Branch_value[keyHex[0]], keyHex[1:])
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	// TODO
	keyHex := keyToHexArr(key)
	if mpt.Root == "" {
		mpt.Values = make(map[string]string)
		mpt.Db = make(map[string]Node)
		keyHex = append(keyHex, uint8(16))
		node := Node{Node_type: 2, Flag_value: Flag_value{compact_encode(keyHex), new_value}}
		hashValue := node.hash_node()
		mpt.Root = hashValue
		mpt.Db[hashValue] = node
	} else {
		mpt.Root = mpt.insert(mpt.Root, keyHex, new_value)
	}
	mpt.Values[key] = new_value
}

func (mpt *MerklePatriciaTrie) insert(hashValue string, keyHex []uint8, value string) string {
	node := mpt.Db[hashValue]
	if node.Node_type == 2 {
		path := compact_decode(node.Flag_value.Encoded_prefix)
		nodeType := node.Flag_value.Encoded_prefix[0] / uint8(16)
		prefixIndex := findPrefixIndex(keyHex, path)
		if nodeType == uint8(0) || nodeType == uint8(1) {
			//Ext.
			return mpt.insertExt(hashValue, prefixIndex, keyHex, path, value)
		}
		//leaf
		return mpt.insertLeaf(hashValue, prefixIndex, keyHex, path, value)
	}
	//branch
	return mpt.insertBranch(hashValue, keyHex, value)
}

func (mpt *MerklePatriciaTrie) insertExt(hashValue string, prefixIndex int, keyHex, path []uint8, value string) string {
	node := mpt.Db[hashValue]
	if prefixIndex == -1 {
		var branch Node
		if len(path) > 1 {
			delete(mpt.Db, hashValue)
			node.Flag_value.Encoded_prefix = compact_encode(path[prefixIndex+2:])
			if len(keyHex) == 0 {
				branch = createBranch(path[0], 16, node.hash_node(), value)
			} else {
				leaf := createLeaf(keyHex[prefixIndex+2:], value)
				branch = createBranch(path[0], keyHex[0], node.hash_node(), leaf.hash_node())
				mpt.Db[leaf.hash_node()] = leaf
			}
			mpt.Db[node.hash_node()] = node
		} else {
			node = mpt.Db[node.Flag_value.Value]
			if len(keyHex) == 0 {
				branch = createBranch(path[0], 16, node.hash_node(), value)
				delete(mpt.Db, hashValue)
			} else {
				leaf := createLeaf(keyHex[prefixIndex+2:], value)
				branch = createBranch(path[0], keyHex[0], node.hash_node(), leaf.hash_node())
				mpt.Db[leaf.hash_node()] = leaf
			}
		}
		mpt.Db[branch.hash_node()] = branch
		return branch.hash_node()
	} else if prefixIndex == len(path)-1 && prefixIndex == len(keyHex)-1 {
		node.Flag_value.Value = mpt.insert(node.Flag_value.Value, keyHex[prefixIndex+1:], value)
		delete(mpt.Db, hashValue)
		mpt.Db[node.hash_node()] = node
		return node.hash_node()
	} else if prefixIndex == len(path)-1 {
		node.Flag_value.Value = mpt.insert(node.Flag_value.Value, keyHex[prefixIndex+1:], value)
		delete(mpt.Db, hashValue)
		mpt.Db[node.hash_node()] = node
		return node.hash_node()
	} else if prefixIndex == len(keyHex)-1 {
		node.Flag_value.Encoded_prefix = compact_encode(path[prefixIndex+2:])
		branch := createBranch(path[prefixIndex+1], 16, node.hash_node(), value)
		ext := createExt(keyHex, branch.hash_node())
		delete(mpt.Db, hashValue)
		mpt.Db[node.hash_node()] = node
		mpt.Db[branch.hash_node()] = branch
		mpt.Db[ext.hash_node()] = ext
		return ext.hash_node()
	} else {
		delete(mpt.Db, hashValue)
		if len(path[prefixIndex+2:]) == 0 {
			node = mpt.Db[node.Flag_value.Value]
		} else {
			node.Flag_value.Encoded_prefix = compact_encode(path[prefixIndex+2:])
			mpt.Db[node.hash_node()] = node
		}
		leaf := createLeaf(keyHex[prefixIndex+2:], value)
		branch := createBranch(path[prefixIndex+1], keyHex[prefixIndex+1], node.hash_node(), leaf.hash_node())
		ext := createExt(path[:prefixIndex+1], branch.hash_node())
		mpt.Db[leaf.hash_node()] = leaf
		mpt.Db[branch.hash_node()] = branch
		mpt.Db[ext.hash_node()] = ext
		return ext.hash_node()
	}
}

func (mpt *MerklePatriciaTrie) insertLeaf(hashValue string, prefixIndex int, keyHex, path []uint8, value string) string {
	node := mpt.Db[hashValue]
	var branch Node
	if prefixIndex == len(path)-1 && prefixIndex == len(keyHex)-1 {
		node.Flag_value.Value = value
		delete(mpt.Db, hashValue)
		mpt.Db[node.hash_node()] = node
		return node.hash_node()
	} else if prefixIndex == -1 {
		lengthPath := len(path)
		lengthKey := len(keyHex)
		delete(mpt.Db, hashValue)
		if lengthKey != 0 && lengthPath != 0 {
			leaf := createLeaf(keyHex[prefixIndex+2:], value)
			index := path[0]
			path = append(path[1:], uint8(16))
			node.Flag_value.Encoded_prefix = compact_encode(path)
			branch = createBranch(index, keyHex[0], node.hash_node(), leaf.hash_node())
			mpt.Db[node.hash_node()] = node
			mpt.Db[leaf.hash_node()] = leaf
		} else if lengthKey == 0 {
			index := path[0]
			path = append(path[1:], uint8(16))
			node.Flag_value.Encoded_prefix = compact_encode(path)
			branch = createBranch(index, 16, node.hash_node(), value)
			mpt.Db[node.hash_node()] = node
		} else {
			leaf := createLeaf(keyHex[prefixIndex+2:], value)
			branch = createBranch(16, keyHex[0], node.Flag_value.Value, leaf.hash_node())
			mpt.Db[leaf.hash_node()] = leaf
		}
		mpt.Db[branch.hash_node()] = branch
		return branch.hash_node()
	} else if prefixIndex == len(path)-1 {
		leaf := createLeaf(keyHex[prefixIndex+2:], value)
		branch = createBranch(keyHex[prefixIndex+1], 16, leaf.hash_node(), node.Flag_value.Value)
		node.Flag_value.Encoded_prefix = compact_encode(path)
		node.Flag_value.Value = branch.hash_node()
		delete(mpt.Db, hashValue)
		mpt.Db[leaf.hash_node()] = leaf
		mpt.Db[branch.hash_node()] = branch
		mpt.Db[node.hash_node()] = node
	} else if prefixIndex == len(keyHex)-1 {
		leaf := createLeaf(path[prefixIndex+2:], node.Flag_value.Value)
		branch = createBranch(path[prefixIndex+1], 16, leaf.hash_node(), value)
		node.Flag_value.Encoded_prefix = compact_encode(path[:prefixIndex+1])
		node.Flag_value.Value = branch.hash_node()
		delete(mpt.Db, hashValue)
		mpt.Db[leaf.hash_node()] = leaf
		mpt.Db[branch.hash_node()] = branch
		mpt.Db[node.hash_node()] = node
	} else {
		oldLeaf := createLeaf(path[prefixIndex+2:], node.Flag_value.Value)
		newLeaf := createLeaf(keyHex[prefixIndex+2:], value)
		branch = createBranch(path[prefixIndex+1], keyHex[prefixIndex+1], oldLeaf.hash_node(), newLeaf.hash_node())
		node.Flag_value.Encoded_prefix = compact_encode(path[:prefixIndex+1])
		node.Flag_value.Value = branch.hash_node()
		delete(mpt.Db, hashValue)
		mpt.Db[oldLeaf.hash_node()] = oldLeaf
		mpt.Db[newLeaf.hash_node()] = newLeaf
		mpt.Db[branch.hash_node()] = branch
		mpt.Db[node.hash_node()] = node
	}
	return node.hash_node()
}

func (mpt *MerklePatriciaTrie) insertBranch(hashValue string, keyHex []uint8, value string) string {
	node := mpt.Db[hashValue]
	if len(keyHex) == 0 {
		node.Branch_value[16] = value
		delete(mpt.Db, hashValue)
		mpt.Db[node.hash_node()] = node
		return node.hash_node()
	} else if node.Branch_value[keyHex[0]] == "" {
		leaf := createLeaf(keyHex[1:], value)
		node.Branch_value[keyHex[0]] = leaf.hash_node()
		delete(mpt.Db, hashValue)
		mpt.Db[leaf.hash_node()] = leaf
		mpt.Db[node.hash_node()] = node
		return node.hash_node()
	} else {
		node.Branch_value[keyHex[0]] = mpt.insert(node.Branch_value[keyHex[0]], keyHex[1:], value)
		delete(mpt.Db, hashValue)
		mpt.Db[node.hash_node()] = node
		return node.hash_node()
	}
}

func createLeaf(keyHex []uint8, value string) Node {
	hexArr := append(keyHex, uint8(16))
	leaf := Node{Node_type: 2, Flag_value: Flag_value{compact_encode(hexArr), value}}
	return leaf
}

func createExt(keyHex []uint8, value string) Node {
	ext := Node{Node_type: 2, Flag_value: Flag_value{compact_encode(keyHex), value}}
	return ext
}

func createBranch(index1, index2 uint8, value1, value2 string) Node {
	var branchValue [17]string
	branchValue[index1] = value1
	branchValue[index2] = value2
	branch := Node{Node_type: 1, Branch_value: branchValue}
	return branch
}

func (mpt *MerklePatriciaTrie) Delete(key string) {
	// TODO
	if mpt.Root != "" {
		keyHex := append(keyToHexArr(key), uint8(16))
		suffix, root := mpt.delete(mpt.Root, keyHex)
		if suffix != nil || root != "" {
			mpt.Root = root
		}
		delete(mpt.Values, key)
	}
}

func (mpt *MerklePatriciaTrie) delete(hashValue string, keyHex []uint8) ([]uint8, string) {
	node := mpt.Db[hashValue]
	if node.Node_type == 2 {
		path := compact_decode(node.Flag_value.Encoded_prefix)
		nodeType := node.Flag_value.Encoded_prefix[0] / uint8(16)
		if nodeType == uint8(0) || nodeType == uint8(1) {
			//Ext
			prefixIndex := findPrefixIndex(keyHex, path)
			return mpt.deleteExt(hashValue, path, keyHex, prefixIndex)
		}
		//leaf
		path = append(path, uint8(16))
		prefixIndex := findPrefixIndex(keyHex, path)
		return mpt.deleteLeaf(hashValue, path, keyHex, prefixIndex)
	}
	//branch
	return mpt.deleteBranch(hashValue, keyHex)
}

func (mpt *MerklePatriciaTrie) deleteExt(hashValue string, path []uint8, keyHex []uint8, prefixIndex int) ([]uint8, string) {
	node := mpt.Db[hashValue]
	if prefixIndex != len(path)-1 {
		return nil, ""
	}
	suffix, value := mpt.delete(node.Flag_value.Value, keyHex[prefixIndex+1:]) // value is the hash of next node
	if suffix != nil {
		delete(mpt.Db, hashValue)
		path = append(path, suffix...)
		node = mpt.Db[value]
		delete(mpt.Db, value)
		node.Flag_value.Encoded_prefix = compact_encode(path)
		mpt.Db[node.hash_node()] = node
		return nil, node.hash_node()
	} else if value != "" {
		delete(mpt.Db, hashValue)
		node.Flag_value.Value = value
		mpt.Db[node.hash_node()] = node
		return nil, node.hash_node()
	}
	return nil, ""
}

func (mpt *MerklePatriciaTrie) deleteLeaf(hashValue string, path []uint8, keyHex []uint8, prefixIndex int) ([]uint8, string) {
	node := mpt.Db[hashValue]
	if prefixIndex == len(path)-1 && prefixIndex == len(keyHex)-1 {
		delete(mpt.Db, hashValue)
		return append(compact_decode(node.Flag_value.Encoded_prefix), uint8(16)), ""
	}
	return nil, ""
}

func (mpt *MerklePatriciaTrie) deleteBranch(hashValue string, keyHex []uint8) ([]uint8, string) {
	node := mpt.Db[hashValue]
	if node.Branch_value[keyHex[0]] == "" {
		return nil, ""
	} else if len(keyHex) == 1 {
		delete(mpt.Db, hashValue)
		if checkElementNum(node.Branch_value) == 2 { //<=2
			return mpt.retrieveNode(node.Branch_value)
		}
		node.Branch_value[16] = ""
		mpt.Db[node.hash_node()] = node
		return nil, node.hash_node()
	}
	suffix, value := mpt.delete(node.Branch_value[keyHex[0]], keyHex[1:])
	if suffix != nil {
		node.Branch_value[keyHex[0]] = value
		delete(mpt.Db, hashValue)
		if value == "" {
			if checkElementNum(node.Branch_value) == 1 {
				if node.Branch_value[16] != "" {
					leaf := createLeaf([]uint8{}, node.Branch_value[16])
					mpt.Db[leaf.hash_node()] = leaf
					return []uint8{uint8(16)}, leaf.hash_node()
				}
				return mpt.retrieveNode(node.Branch_value)
			}
		}
		mpt.Db[node.hash_node()] = node
		return nil, node.hash_node()
	} else if value != "" {
		delete(mpt.Db, hashValue)
		node.Branch_value[keyHex[0]] = value
		mpt.Db[node.hash_node()] = node
		return nil, node.hash_node()
	}
	return nil, ""
}

func (mpt *MerklePatriciaTrie) retrieveNode(branchValue [17]string) ([]uint8, string) {
	for i := 0; i < 16; i++ {
		if branchValue[i] != "" {
			node := mpt.Db[branchValue[i]]
			if node.Node_type == 2 {
				delete(mpt.Db, branchValue[i])
				noteType := node.Flag_value.Encoded_prefix[0] / uint8(16)
				suffix := append([]uint8{uint8(i)}, compact_decode(node.Flag_value.Encoded_prefix)...)
				if noteType == uint8(2) || noteType == uint8(3) {
					leaf := createLeaf(suffix, node.Flag_value.Value)
					mpt.Db[leaf.hash_node()] = leaf
					return append(suffix, uint8(16)), leaf.hash_node()
				}
				ext := createExt(suffix, node.Flag_value.Value)
				mpt.Db[ext.hash_node()] = ext
				return suffix, ext.hash_node()
			}
			ext := createExt([]uint8{uint8(i)}, branchValue[i])
			mpt.Db[ext.hash_node()] = ext
			return []uint8{uint8(i)}, ext.hash_node()
		}
	}
	return nil, ""
}

func checkElementNum(branchValue [17]string) int {
	i := 0
	for _, v := range branchValue {
		if v != "" {
			i++
		}
	}
	return i
}

func keyToHexArr(key string) []uint8 {
	var keyHex []uint8
	for i := 0; i < len(key); i++ {
		keyHex = append(keyHex, key[i]/uint8(16))
		keyHex = append(keyHex, key[i]%uint8(16))
	}
	return keyHex
}

func findPrefixIndex(keyHex []uint8, path []uint8) int {
	i := 0
	for i < min(len(keyHex), len(path)) {
		if keyHex[i] != path[i] {
			break
		}
		i++
	}
	return i - 1
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func compact_encode(hex_array []uint8) []uint8 {
	// TODO
	var term uint8 = 0
	if hex_array[len(hex_array)-1] == 16 {
		hex_array = hex_array[:len(hex_array)-1]
		term = 1
	}
	oddlen := uint8(len(hex_array) % 2)
	flags := 2*term + oddlen
	if oddlen == 1 {
		hex_array = append([]uint8{flags}, hex_array...)
	} else {
		hex_array = append([]uint8{flags, uint8(0)}, hex_array...)
	}
	var encodedArr []uint8
	for i := 0; i < len(hex_array); i = i + 2 {
		encodedArr = append(encodedArr, uint8(16)*hex_array[i]+hex_array[i+1])
	}
	return encodedArr
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	// TODO
	var decodedArr []uint8
	for i := 0; i < len(encoded_arr); i++ {
		decodedArr = append(decodedArr, encoded_arr[i]/uint8(16))
		decodedArr = append(decodedArr, encoded_arr[i]%uint8(16))
	}
	if decodedArr[0] == uint8(0) || decodedArr[0] == uint8(2) {
		decodedArr = decodedArr[2:]
	} else {
		decodedArr = decodedArr[1:]
	}
	return decodedArr
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func (node *Node) hash_node() string {
	var str string
	switch node.Node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.Branch_value {
			str += v
		}
	case 2:
		str = node.Flag_value.Value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.Node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.Branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.Branch_value[16])
	case 2:
		Encoded_prefix := node.Flag_value.Encoded_prefix
		node_name := "Leaf"
		if is_ext_node(Encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(Encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.Flag_value.Value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.Db = make(map[string]Node)
	mpt.Root = ""
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.Db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.Db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}
