package proj2

// CS 161 Project 2 Spring 2020
// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder. We will be very upset.

import (
	// You neet to add with
	// go get github.com/cs161-staff/userlib

	"github.com/cs161-staff/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...
	"encoding/json"

	// Likewise useful for debugging, etc...
	"encoding/hex"

	// UUIDs are generated right based on the cryptographic PRNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"

	// Useful for debug messages, or string manipulation for datastore keys.
	"strings"

	// Want to import errors.
	"errors"

	// Optional. You can remove the "_" there, but please do not touch
	// anything else within the import bracket.
	_ "strconv"
	// if you are looking for fmt, we don't give you fmt, but you can use userlib.DebugMsg.
	// see someUsefulThings() below:
)

// This serves two purposes:
// a) It shows you some useful primitives, and
// b) it suppresses warnings for items not being imported.
// Of course, this function can be deleted.
func someUsefulThings() {
	// Creates a random UUID
	f := uuid.New()
	userlib.DebugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	userlib.DebugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	userlib.DebugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	userlib.DebugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	userlib.DebugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	userlib.DebugMsg("Creation of error %v", errors.New(strings.ToTitle("This is an error")))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var pk userlib.PKEEncKey
	var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("Key is %v, %v", pk, sk)
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

//File struct contains file data
type File struct {
	Data [][]byte
	FMac [][]byte
}

//FileInfo struct describes a stored file
type FileInfo struct {
	Fuuid  uuid.UUID
	FCsalt []byte
}

// User structure defines a user record
type User struct {
	Username string

	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
	SymKey    []byte
	Private   userlib.PKEDecKey
	Signature userlib.DSSignKey
	Files     map[string]uuid.UUID
}

//Wrap struct is a wrapper containing a user instance and an auth struct
type Wrap struct {
	Salt1   []byte
	Salt2   []byte
	Hashed  []byte
	EncUser []byte //Encrypted User struct
}

func isEqual(one []byte, two []byte) bool {
	if len(one) != len(two) {
		return false
	}
	for i := range one {
		if one[i] != two[i] {
			return false
		}
	}
	return true
}

func bHashKDF(key []byte, purpose string) []byte {
	userkey, _ := userlib.HashKDF(key[:16], []byte(purpose))
	return userkey[:16]
}

func dUser(username string) (wrapper Wrap, err error) {
	wuuid, _ := uuid.FromBytes(userlib.Argon2Key([]byte(username), []byte(username), uint32(16)))
	mwrapper, found := userlib.DatastoreGet(wuuid)
	if !found {
		return wrapper, errors.New(strings.ToTitle("Username is invalid!"))
	}
	json.Unmarshal(mwrapper, &wrapper)
	return wrapper, nil
}
func dinfo(infouuid uuid.UUID) (fileinfo FileInfo, err error) {
	minfo, ok := userlib.DatastoreGet(infouuid)
	if !ok {
		return fileinfo, errors.New(strings.ToTitle("File Info not found!"))
	}
	err = json.Unmarshal(minfo, &fileinfo)
	if err != nil {
		return fileinfo, err
	}
	return fileinfo, nil
}
func dfile(fuuid uuid.UUID) (file File, err error) {
	mfile, ok := userlib.DatastoreGet(fuuid)
	if !ok {
		return file, errors.New(strings.ToTitle("File not found!"))
	}
	err = json.Unmarshal(mfile, &file)
	if err != nil {
		return file, err
	}
	return file, nil
}

func kUser(username string) (value userlib.PKEEncKey, err error) {
	value, found := userlib.KeystoreGet(username)
	if !found {
		return value, errors.New(strings.ToTitle("Username is invalid!"))
	}
	return value, nil
}

func (userdata *User) refreshUser() error {
	wrapper, err := dUser(userdata.Username)
	if err != nil {
		return err
	}
	userkey := bHashKDF(userdata.SymKey, "user")
	muserdata := userlib.SymDec(userkey, wrapper.EncUser)
	err = json.Unmarshal(muserdata, &userdata)
	if err != nil {
		return err
	}
	return nil
}

func (userdata *User) setUser(wrapper Wrap) {
	muserdata, _ := json.Marshal(userdata)
	userkey := bHashKDF(userdata.SymKey, "user")
	encuser := userlib.SymEnc(userkey, userlib.RandomBytes(16), muserdata)
	wrapper.EncUser = encuser

	mwrapper, _ := json.Marshal(wrapper)
	wuuid, _ := uuid.FromBytes(userlib.Argon2Key([]byte(userdata.Username), []byte(userdata.Username), uint32(16)))
	userlib.DatastoreSet(wuuid, mwrapper)
}
func (userdata *User) updateUser() error {
	wrapper, err := dUser(userdata.Username)
	if err != nil {
		return err
	}
	userdata.setUser(wrapper)
	return nil
}

func (userdata *User) allFile(filename string) (fileinfo FileInfo, file File, fKey []byte, err error) {
	infouuid, ok := userdata.Files[filename]
	if !ok {
		return fileinfo, file, fKey, errors.New(strings.ToTitle("File not found!"))
	}
	fileinfo, err = dinfo(infouuid)
	if err != nil {
		return fileinfo, file, fKey, err
	}
	file, err = dfile(fileinfo.Fuuid)
	if err != nil {
		return fileinfo, file, fKey, err
	}
	fKey = userlib.Argon2Key([]byte(filename), fileinfo.FCsalt, uint32(32))
	return fileinfo, file, fKey, nil
}

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the password has strong entropy, EXCEPT
// the attackers may possess a precomputed tables containing
// hashes of common passwords downloaded from the internet.
func InitUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	var wrapper Wrap
	userdataptr = &userdata

	wrapper.Salt1 = userlib.RandomBytes(32)
	wrapper.Salt2 = userlib.RandomBytes(32)
	wrapper.Hashed = userlib.Argon2Key([]byte(password), wrapper.Salt1, uint32(32))

	public, private, _ := userlib.PKEKeyGen()
	sign, verify, _ := userlib.DSKeyGen()
	err = userlib.KeystoreSet(string(username+"_enc"), public)
	if err != nil {
		return nil, err
	}
	err = userlib.KeystoreSet(string(username+"_sig"), verify)
	if err != nil {
		return nil, err
	}

	//TODO: This is a toy implementation.
	userdata.Username = username
	userdata.SymKey = userlib.Argon2Key([]byte(password), wrapper.Salt2, uint32(32))
	userdata.Private = private
	userdata.Signature = sign
	userdata.Files = make(map[string]uuid.UUID)
	//End of toy implementation
	userdataptr.setUser(wrapper)

	return userdataptr, nil
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata

	wrapper, err := dUser(username)
	if err != nil {
		return nil, err
	}

	candidate := userlib.Argon2Key([]byte(password), wrapper.Salt1, uint32(32))
	if !isEqual(candidate, wrapper.Hashed) {
		return nil, errors.New(strings.ToTitle("Password is incorrect!"))
	}

	userdata.SymKey = userlib.Argon2Key([]byte(password), wrapper.Salt2, uint32(32))
	userdata.Username = username
	err = userdataptr.refreshUser()
	if err != nil {
		return nil, err
	}
	return userdataptr, nil
}

// This stores a file in the datastore.
//
// The plaintext of the filename + the plaintext and length of the filename
// should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {
	userdata.refreshUser()

	var fileinfo FileInfo
	var file File

	fileinfo.FCsalt = userlib.RandomBytes(32)
	fKey := userlib.Argon2Key([]byte(filename), fileinfo.FCsalt, uint32(32))
	fileinfo.Fuuid, _ = uuid.FromBytes(bHashKDF(fKey, "fuuid"))

	//TODO: This is a toy implementation.
	encdata := userlib.SymEnc(bHashKDF(fKey, "file"), userlib.RandomBytes(16), data)
	fmac, _ := userlib.HMACEval(bHashKDF(fKey, "fmac"), encdata)
	file.FMac = append(file.FMac, fmac)

	file.Data = append(file.Data, encdata)
	mfile, _ := json.Marshal(file)
	var file2 File
	json.Unmarshal(mfile, &file2)
	userlib.DatastoreSet(fileinfo.Fuuid, mfile)
	//End of toy implementation

	infouuid, _ := uuid.FromBytes(bHashKDF(fKey, "infouuid"))
	userdata.Files[filename] = infouuid
	minfo, _ := json.Marshal(fileinfo)
	userlib.DatastoreSet(infouuid, minfo)

	userdata.updateUser()

	return
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.
func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	userdata.refreshUser()
	fileinfo, file, fKey, err := userdata.allFile(filename)
	if err != nil {
		return err
	}

	encdata := userlib.SymEnc(bHashKDF(fKey, "file"), userlib.RandomBytes(16), data)
	fmac, _ := userlib.HMACEval(bHashKDF(fKey, "fmac"), encdata)
	file.FMac = append(file.FMac, fmac)

	file.Data = append(file.Data, encdata)
	mfile, _ := json.Marshal(file)
	userlib.DatastoreSet(fileinfo.Fuuid, mfile)
	userdata.updateUser()

	return nil
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	userdata.refreshUser()
	_, file, fKey, err := userdata.allFile(filename)
	if err != nil {
		return nil, err
	}
	for i, datapart := range file.Data {
		fmac, _ := userlib.HMACEval(bHashKDF(fKey, "fmac"), datapart)
		if !userlib.HMACEqual(fmac, file.FMac[i]) {
			return nil, errors.New(strings.ToTitle("File is corrupted"))
		}
		decdata := userlib.SymDec(bHashKDF(fKey, "file"), datapart)
		data = append(data, decdata...)
	}
	return data, nil
}

// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.
func (userdata *User) ShareFile(filename string, recipient string) (
	magic_string string, err error) {

	return
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {
	return nil
}

// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	return
}
