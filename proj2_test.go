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
	"github.com/google/uuid"
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

//test two users with same username
func TestSameUsernameInit(t *testing.T) {
	clear()
	t.Log("Initializing two users with same username")
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	_, err = InitUser("alice", "acorn")
	if err == nil {
		t.Error("Failed to catch redundant username")
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

func TestAdvGet(t *testing.T) {
	clear()
	t.Log("Get test")

	alice, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	malice, err := InitUser("malice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	if reflect.DeepEqual(alice, malice) {
		t.Error("Init returned the same values for alice and malice")
		return
	}

	m := userlib.DatastoreGetMap()
	var uuids []uuid.UUID
	for k := range m {
		uuids = append(uuids, k)
	}

	one, _ := userlib.DatastoreGet(uuids[0])
	two, _ := userlib.DatastoreGet(uuids[1])
	if reflect.DeepEqual(one, two) {
		t.Error("Stored entries are already equal???")
	}
	userlib.DatastoreSet(uuids[1], one)
	userlib.DatastoreSet(uuids[0], two)

	u2, err := GetUser("alice", "fubar")
	if err == nil {
		t.Error("Failed to detect corruption", u2)
		return
	}
}

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

//test second instantiation store and edit
func TestSameAppend(t *testing.T) {
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
	u3, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to get user again", err)
		return
	}
	v := []byte("This is a test")
	u1.StoreFile("file1", v)
	vmore := []byte("!")
	u2.AppendFile("file1", vmore)
	u3.AppendFile("file1", vmore)
	final := []byte("This is a test!!")
	v2, err2 := u1.LoadFile("file1")
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

//test modify file after sharing
func TestshareAppend(t *testing.T) {
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

	var magic_string string

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
	apple := []byte("!")
	u2.AppendFile("file2", apple)
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
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

//test share to dummy user
func TestRevokeDummy(t *testing.T) {
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
	u3, err3 := InitUser("dummy", "foo")
	if err3 != nil {
		t.Error("Failed to initialize dummy", err3)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var magic_string string
	var magic_string2 string

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
	magic_string2, err = u2.ShareFile("file2", "dummy")
	if err != nil {
		t.Error("Failed to receive the share message to dummy", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string2)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	err = u.RevokeFile("file1", "bob")
	if err != nil {
		t.Error("Failed to revoke file", err)
	}

	_, err = u3.LoadFile("file3")
	if err == nil {
		t.Error("Still was able to download the file after revoke")
		return
	}
	v3 := []byte("This is a test")
	err = u3.AppendFile("file3", v3)
	if err == nil {
		t.Error("still updates after revoke", err)
		return
	}
}

//revoke called by nonoriginal author
func TestNonOgRevoke(t *testing.T) {
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
	u3, err3 := InitUser("dummy", "foo")
	if err3 != nil {
		t.Error("Failed to initialize dummy", err3)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var magic_string string
	var magic_string2 string

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
	magic_string2, err = u2.ShareFile("file2", "dummy")
	if err != nil {
		t.Error("Failed to receive the share message to dummy", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string2)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	err = u2.RevokeFile("file2", "dummy")
	if err == nil {
		t.Error("u2 shouldn't be able to revoke", err)
	}

	_, err = u3.LoadFile("file3")
	if err != nil {
		t.Error("should have access")
		return
	}
	v3 := []byte("This is a test")
	err = u3.AppendFile("file3", v3)
	if err != nil {
		t.Error("should still update after revoke", err)
		return
	}
}

//other non revoked users try to update file/load file
func TestMultiRevoke(t *testing.T) {
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
	u3, err3 := InitUser("cat", "c")
	if err3 != nil {
		t.Error("Failed to initialize cat")
		return
	}
	u4, err4 := InitUser("dummy", "d")
	if err4 != nil {
		t.Error("Failed to initialized dummy")
		return
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	var magic_string string
	var magic_string2 string
	var magic_string3 string
	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("sharing failed bob", err)
		return
	}
	magic_string2, err = u.ShareFile("file1", "cat")
	if err != nil {
		t.Error("sharing failed cat", err)
		return
	}
	magic_string3, err = u.ShareFile("file1", "dummy")
	if err != nil {
		t.Error("sharing failed dummy", err)
	}
	err = u2.ReceiveFile("file1", "alice", magic_string)
	if err != nil {
		t.Error("bob Failed to receive the share message", err)
		return
	}
	err = u3.ReceiveFile("file1", "alice", magic_string2)
	if err != nil {
		t.Error("cat Failed to receive", err)
		return
	}
	err = u4.ReceiveFile("file1", "alice", magic_string3)
	if err != nil {
		t.Error("dummy failed to receive", err)
		return
	}
	err = u.RevokeFile("file1", "dummy")
	if err != nil {
		t.Error("failed to revoke dummy", err)
		return
	}
	v3 := []byte("!")
	err = u2.AppendFile("file1", v3)
	if err != nil {
		t.Error("bob can't append", err)
		return
	}
	catfile, errload := u3.LoadFile("file1")
	if errload != nil {
		t.Error("cat should have access", err)
		return
	}
	ogfile, ogerr := u.LoadFile("file1")
	if ogerr != nil {
		t.Error("alice can't load", err)
		return
	}
	eq := reflect.DeepEqual(ogfile, catfile)
	if !eq {
		t.Error("file not synchronized")
		return
	}
}

//magic string gets modified
func TestMagicMod(t *testing.T) {
	clear()
	t.Log("Magic string modification test")
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

	var magic_string string
	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	magic_string = "aaaa"
	err = u2.ReceiveFile("file1", "alice", magic_string)
	if err == nil {
		t.Error("Failed to catch wrong magic_string", err)
		return
	}
}

//empty filename
func TestEmptyFileName(t *testing.T) {
	clear()
	t.Log("Empty Filename test")
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	v := []byte("This is a test")
	u.StoreFile("", v)
	var download []byte
	download, err = u.LoadFile("")
	eq := reflect.DeepEqual(download, v)
	if !eq {
		t.Error("fails to load empty file")
		return
	}
}
func TestRevoke(t *testing.T) {
	clear()
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
	err = u.RevokeFile("file1", "bob")
	if err != nil {
		t.Error("Failed to revoke file", err)
	}

	_, err = u2.LoadFile("file2")
	if err == nil {
		t.Error("Still was able to download the file after revoke")
		return
	}
}
