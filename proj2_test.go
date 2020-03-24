package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	_ "encoding/hex"
	_ "encoding/json"
	_ "errors"
	"reflect"
	_ "strconv"
	_ "strings"
	"testing"

	"github.com/cs161-staff/userlib"
	_ "github.com/google/uuid"
)

func clear() {
	// Wipes the storage so one test does not affect another
	userlib.DatastoreClear()
	userlib.KeystoreClear()
}

func TestInit(t *testing.T) {
	clear()
	t.Log("Initialization test")

	// You can set this to false!
	userlib.SetDebugStatus(true)

	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	// t.Log() only produces output if you run with "go test -v"
	t.Log("Got user", u)
	// If you want to comment the line above,
	// write _ = u here to make the compiler happy
	// You probably want many more tests here.
}

func TestDoubleInit(t *testing.T) {
	clear()
	t.Log("Initializing User twice test")
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	_, err = InitUser("alice", "fubar")
	if err == nil {
		t.Error("Failed to catch reinitialization")
		return
	}

	_, err = InitUser("malice", "fubar")
	if err != nil {
		t.Error("Unique users can have the same password", err)
		return
	}
}

func TestInvalidInit(t *testing.T) {
	clear()
	t.Log("Initializing Invalid User")
	_, err := InitUser("", "")
	if err == nil {
		t.Error("Failed to catch empty username and password")
		return
	}
	_, err = InitUser("alice", "")
	if err == nil {
		t.Error("Failed to catch empty password")
		return
	}
	_, err = InitUser("", "fubar")
	if err == nil {
		t.Error("Failed to catch empty username")
		return
	}
}

func TestGet(t *testing.T) {
	clear()
	t.Log("Get test")

	// You can set this to false!
	userlib.SetDebugStatus(true)

	u1, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to get user", err)
		return
	}
	equiv := reflect.DeepEqual(u1, u2)

	if !equiv {
		t.Error("Failed to get correct user", err)
	}

	t.Log("Get works!", u2)
}

func TestNilGet(t *testing.T) {
	clear()
	t.Log("Get nonexistent user test")
	_, err := GetUser("alice", "fubar")
	if err == nil {
		t.Error("Failed to error")
		return
	}
}

func TestInvalidGet(t *testing.T) {
	clear()
	t.Log("Get test")

	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	_, err = GetUser("", "")
	if err == nil {
		t.Error("Failed to catch empty username and password")
		return
	}
	_, err = GetUser("alice", "")
	if err == nil {
		t.Error("Failed to catch empty password")
		return
	}
	_, err = GetUser("", "fubar")
	if err == nil {
		t.Error("Failed to catch empty username")
		return
	}
}

// func TestAdvGet(t *testing.T) {
// 	clear()
// 	t.Log("Get test")

// 	alice, err := InitUser("alice", "fubar")
// 	if err != nil {
// 		// t.Error says the test fails
// 		t.Error("Failed to initialize user", err)
// 		return
// 	}

// 	malice, err := InitUser("malice", "fubar")
// 	if err != nil {
// 		// t.Error says the test fails
// 		t.Error("Failed to initialize user", err)
// 		return
// 	}

// 	if reflect.DeepEqual(alice, malice) {
// 		t.Error("Init returned the same values for alice and malice")
// 		return
// 	}

// 	m := userlib.DatastoreGetMap()
// 	var uuids []uuid.UUID
// 	for k := range m {
// 		uuids = append(uuids, k)
// 	}

// 	one, _ := userlib.DatastoreGet(uuids[0])
// 	two, _ := userlib.DatastoreGet(uuids[1])
// 	if reflect.DeepEqual(one, two) {
// 		t.Error("Stored entries are already equal???")
// 	}
// 	userlib.DatastoreSet(uuids[1], one)
// 	userlib.DatastoreSet(uuids[0], two)

// 	u2, err := GetUser("alice", "fubar")
// 	if err == nil {
// 		t.Error("Failed to detect corruption", u2)
// 		return
// 	}
// }

func TestStorage(t *testing.T) {
	clear()
	t.Log("Storage test")
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestAppend(t *testing.T) {
	clear()
	t.Log("Append test")
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	vmore := []byte("!")
	u.AppendFile("file1", vmore)
	final := []byte("This is a test!")
	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(final, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestInvalidFile(t *testing.T) {
	clear()
	t.Log("Invalid File test")
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	_, err2 := u.LoadFile("this file does not exist")
	if err2 == nil {
		t.Error("Downloaded a ninexistent file", err2)
		return
	}
}

func TestShare(t *testing.T) {
	clear()
	t.Log("Share test")
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

}
