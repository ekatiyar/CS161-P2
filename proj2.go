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

//Helper function: Conversts uuid to []byte type
func uuidtobytes(id uuid.UUID) (data []byte, err error) {
	return json.Marshal(id)
}

//File struct contains file data
type File struct {
	Data [][]byte
	Mac  [][]byte
}

//FileInfo struct describes a stored file
type FileInfo struct {
	Fuuid uuid.UUID
	FKey  []byte
	Owner bool
}

//InfoWrapper struct describes a FileInfo struct
type InfoWrapper struct {
	Infouuid uuid.UUID
	InfoKey  []byte //Decryptes FileInfo struct
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
	FilesMap  uuid.UUID
	//for sharing
}

//Wrap struct is a wrapper containing a user instance and an auth struct
type Wrap struct {
	Salt    []byte
	EncUser []byte //Encrypted User struct
}

//Share struct is used for sharing
type Share struct {
	EncInfoWrapper []byte //Encrypted using recipient's public key
	Sign           []byte
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

func kUser(username string) (value userlib.PKEEncKey, err error) {
	value, found := userlib.KeystoreGet(string(username + "_enc"))
	if !found {
		return value, errors.New(strings.ToTitle("Username is invalid!"))
	}
	return value, nil
}

func kVerify(username string) (value userlib.DSVerifyKey, err error) {
	value, found := userlib.KeystoreGet(string(username + "_sig"))
	if !found {
		return value, errors.New(strings.ToTitle("Username is invalid!"))
	}
	return value, nil
}

// bHashKDF runs HashKDF but outputs a 16-byte long slice
func bHashKDF(key []byte, purpose string) []byte {
	userkey, _ := userlib.HashKDF(key[:16], []byte(purpose))
	return userkey[:16]
}

// dUser returns a wrapper given a username
func dUser(username string) (wrapper Wrap, err error) {
	wuuid, _ := uuid.FromBytes(userlib.Argon2Key([]byte(username), nil, uint32(16)))
	mwrapper, found := userlib.DatastoreGet(wuuid)
	if !found {
		return wrapper, errors.New(strings.ToTitle("Username not found!"))
	}
	err = json.Unmarshal(mwrapper, &wrapper)
	if err != nil {
		return wrapper, err
	}
	return wrapper, nil
}

// updateFilesMap updates the FilesMap for a new file or a new user to the file
func (userdata *User) updateFilesMap(filename string, infowrapper InfoWrapper, username string) (err error) {
	var filesmap map[string]map[string]InfoWrapper
	encmfilesmap, ok := userlib.DatastoreGet(userdata.FilesMap)
	if !ok {
		return errors.New(strings.ToTitle("Unable to load filesmap"))
	}
	mfilesmap := userlib.SymDec(bHashKDF(userdata.SymKey, "fmap"), encmfilesmap)
	err = json.Unmarshal(mfilesmap, &filesmap)
	if err != nil {
		return err
	}

	_, ok = filesmap[filename]
	if !ok {
		filesmap[filename] = map[string]InfoWrapper{}
	}
	filesmap[filename][username] = infowrapper

	mfilesmap, _ = json.Marshal(filesmap)
	encmfilesmap = userlib.SymEnc(bHashKDF(userdata.SymKey, "fmap"), userlib.RandomBytes(16), mfilesmap)
	userlib.DatastoreSet(userdata.FilesMap, encmfilesmap)
	return nil
}

func (userdata *User) allFile(filename string) (infowrapper InfoWrapper, fileinfo FileInfo, file File, err error) {
	var filesmap map[string]map[string]InfoWrapper
	encmfilesmap, ok := userlib.DatastoreGet(userdata.FilesMap)
	if !ok {
		return infowrapper, fileinfo, file, errors.New(strings.ToTitle("InfoWrapper not found"))
	}
	mfilesmap := userlib.SymDec(bHashKDF(userdata.SymKey, "fmap"), encmfilesmap)
	err = json.Unmarshal(mfilesmap, &filesmap)
	if err != nil {
		return infowrapper, fileinfo, file, err
	}

	fileaccess, ok := filesmap[filename]
	if !ok {
		return infowrapper, fileinfo, file, errors.New(strings.ToTitle("File not found!"))
	}
	infowrapper, ok = fileaccess[userdata.Username]
	if !ok {
		return infowrapper, fileinfo, file, errors.New(strings.ToTitle(userdata.Username + " does not have access to this file"))
	}

	encminfo, ok := userlib.DatastoreGet(infowrapper.Infouuid)
	if !ok {
		return infowrapper, fileinfo, file, errors.New(strings.ToTitle("File Info not found!"))
	}
	minfo := userlib.SymDec(infowrapper.InfoKey, encminfo)
	err = json.Unmarshal(minfo, &fileinfo)
	if err != nil {
		return infowrapper, fileinfo, file, err
	}

	mfile, ok := userlib.DatastoreGet(fileinfo.Fuuid)
	if !ok {
		return infowrapper, fileinfo, file, errors.New(strings.ToTitle("File not found!"))
	}
	err = json.Unmarshal(mfile, &file)
	if err != nil {
		return infowrapper, fileinfo, file, err
	}
	return infowrapper, fileinfo, file, nil
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
	if username == "" || password == "" {
		return userdataptr, errors.New(strings.ToTitle("Username and Password must not be empty"))
	}
	_, err = dUser(username)
	if err == nil {
		return userdataptr, errors.New(strings.ToTitle("Username already exists!"))
	}

	var userdata User
	var wrapper Wrap
	userdataptr = &userdata

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
	wrapper.Salt = userlib.RandomBytes(16)

	userdata.Username = username
	userdata.SymKey = userlib.Argon2Key([]byte(username+password), wrapper.Salt, uint32(16))
	userdata.Private = private
	userdata.Signature = sign
	userdata.FilesMap = uuid.New()

	filesmap := map[string]map[string]InfoWrapper{}
	mfilesmap, _ := json.Marshal(filesmap)
	encmfilesmap := userlib.SymEnc(bHashKDF(userdata.SymKey, "fmap"), userlib.RandomBytes(16), mfilesmap)
	userlib.DatastoreSet(userdata.FilesMap, encmfilesmap)

	muserdata, _ := json.Marshal(userdata)
	userkey := bHashKDF(userdata.SymKey, "user")
	encuser := userlib.SymEnc(userkey, userlib.RandomBytes(16), muserdata)

	wrapper.EncUser = encuser
	mwrapper, _ := json.Marshal(wrapper)
	wuuid, _ := uuid.FromBytes(userlib.Argon2Key([]byte(userdata.Username), nil, uint32(16)))
	userlib.DatastoreSet(wuuid, mwrapper)

	return userdataptr, nil
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	if username == "" || password == "" {
		return userdataptr, errors.New(strings.ToTitle("Username and Password must not be empty"))
	}

	var userdata User
	userdataptr = &userdata

	wrapper, err := dUser(username)
	if err != nil {
		return nil, err
	}

	userkey := bHashKDF(userlib.Argon2Key([]byte(username+password), wrapper.Salt, uint32(16)), "user")
	muserdata := userlib.SymDec(userkey, wrapper.EncUser)
	err = json.Unmarshal(muserdata, &userdata)
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
	_, fileinfo, _, err := userdata.allFile(filename)
	if err != nil {
		fileinfo.FKey = userlib.Argon2Key([]byte(filename), userlib.RandomBytes(16), uint32(16))
		fileinfo.Fuuid = uuid.New()
		fileinfo.Owner = true
		infouuid := uuid.New()
		minfo, _ := json.Marshal(fileinfo)
		bytes, _ := uuidtobytes(infouuid)
		InfoKey := bHashKDF(userdata.SymKey, string(bytes))
		encinfo := userlib.SymEnc(InfoKey, userlib.RandomBytes(16), minfo)
		userlib.DatastoreSet(infouuid, encinfo)

		var temp InfoWrapper
		temp.Infouuid = infouuid
		temp.InfoKey = InfoKey
		err = userdata.updateFilesMap(filename, temp, userdata.Username)
		if err != nil {
			return
		}
	}

	var file File

	encdata := userlib.SymEnc(bHashKDF(fileinfo.FKey, "file"), userlib.RandomBytes(16), data)
	file.Data = append(file.Data, encdata)
	fmac, _ := userlib.HMACEval(bHashKDF(fileinfo.FKey, "mac"), encdata)
	file.Mac = append(file.Mac, fmac)

	mfile, _ := json.Marshal(file)
	userlib.DatastoreSet(fileinfo.Fuuid, mfile)
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.
func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	_, fileinfo, file, err := userdata.allFile(filename)
	if err != nil {
		return err
	}

	encdata := userlib.SymEnc(bHashKDF(fileinfo.FKey, "file"), userlib.RandomBytes(16), data)
	file.Data = append(file.Data, encdata)
	fmac, _ := userlib.HMACEval(bHashKDF(fileinfo.FKey, "mac"), encdata)
	file.Mac = append(file.Mac, fmac)

	mfile, _ := json.Marshal(file)
	userlib.DatastoreSet(fileinfo.Fuuid, mfile)

	return nil
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	_, fileinfo, file, err := userdata.allFile(filename)
	if err != nil {
		return nil, err
	}

	for i, datapart := range file.Data {
		fmac, _ := userlib.HMACEval(bHashKDF(fileinfo.FKey, "mac"), datapart)
		if !userlib.HMACEqual(fmac, file.Mac[i]) {
			return nil, errors.New(strings.ToTitle("File is corrupted"))
		}
		decdata := userlib.SymDec(bHashKDF(fileinfo.FKey, "file"), datapart)
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
	//magic_string is a marshaled struct with uuid hashed and converted to string, signature
	ek, err := kUser(recipient)
	if err != nil {
		return magic_string, nil
	}
	infowrapper, fileinfo, _, err := userdata.allFile(filename)
	if err == nil {
		if fileinfo.Owner == true {
			fileinfo.Owner = false
			infowrapper.Infouuid = uuid.New()
			minfo, _ := json.Marshal(fileinfo)
			bytes, err := uuidtobytes(infowrapper.Infouuid)
			if err != nil {
				return magic_string, err
			}
			InfoKey := bHashKDF(userdata.SymKey, string(bytes))
			encinfo := userlib.SymEnc(InfoKey, userlib.RandomBytes(16), minfo)
			userlib.DatastoreSet(infowrapper.Infouuid, encinfo)

			infowrapper.InfoKey = InfoKey
			err = userdata.updateFilesMap(filename, infowrapper, recipient)
			if err != nil {
				return magic_string, err
			}
		}

		var share Share
		minfowrapper, _ := json.Marshal(infowrapper)
		encinfowrapper, err := userlib.PKEEnc(ek, minfowrapper)
		if err != nil {
			return magic_string, err
		}
		share.EncInfoWrapper = encinfowrapper
		signature, _ := userlib.DSSign(userdata.Signature, encinfowrapper)
		share.Sign = signature

		mshare, _ := json.Marshal(share)
		return string(mshare), nil

	} else {
		return magic_string, err
	}
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {
	var share Share
	err := json.Unmarshal([]byte(magic_string), &share)
	if err != nil {
		return err
	}
	vk, err := kVerify(sender)
	if err != nil {
		return err
	}
	err = userlib.DSVerify(vk, share.EncInfoWrapper, share.Sign)
	if err != nil {
		return err
	}
	minfowrapper, err := userlib.PKEDec(userdata.Private, share.EncInfoWrapper)
	if err != nil {
		return err
	}
	var infowrapper InfoWrapper
	err = json.Unmarshal(minfowrapper, &infowrapper)
	if err != nil {
		return err
	}
	err = userdata.updateFilesMap(filename, infowrapper, userdata.Username)
	if err != nil {
		return err
	}
	return nil
}

func (userdata *User) mapRevoke(filename string, target_username string) (newWrapperMap map[string]InfoWrapper, err error) {
	var filesmap map[string]map[string]InfoWrapper
	encmfilesmap, ok := userlib.DatastoreGet(userdata.FilesMap)
	if !ok {
		return newWrapperMap, errors.New(strings.ToTitle("Filesmap not found!"))
	}
	mfilesmap := userlib.SymDec(bHashKDF(userdata.SymKey, "fmap"), encmfilesmap)
	err = json.Unmarshal(mfilesmap, &filesmap)
	if err != nil {
		return newWrapperMap, err
	}

	newWrapperMap, ok = filesmap[filename]
	if !ok {
		filesmap[filename] = map[string]InfoWrapper{}
	}
	targwrapper, ok := newWrapperMap[target_username]
	if !ok {
		return newWrapperMap, errors.New(strings.ToTitle(target_username + " does not have access to this file"))
	}
	userlib.DatastoreDelete(targwrapper.Infouuid)
	delete(newWrapperMap, target_username)
	filesmap[filename] = newWrapperMap

	mfilesmap, _ = json.Marshal(filesmap)
	encmfilesmap = userlib.SymEnc(bHashKDF(userdata.SymKey, "fmap"), userlib.RandomBytes(16), mfilesmap)
	userlib.DatastoreSet(userdata.FilesMap, encmfilesmap)
	return newWrapperMap, nil
}

// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	infowrapper, fileinfo, file, err := userdata.allFile(filename)
	if err != nil {
		return err
	}
	data, err := userdata.LoadFile(filename)
	if err != nil {
		return err
	}

	fileinfo.FKey = userlib.Argon2Key([]byte(filename), userlib.RandomBytes(16), uint32(16))
	minfo, _ := json.Marshal(fileinfo)
	encinfo := userlib.SymEnc(infowrapper.InfoKey, userlib.RandomBytes(16), minfo)
	userlib.DatastoreSet(infowrapper.Infouuid, encinfo)

	encdata := userlib.SymEnc(bHashKDF(fileinfo.FKey, "file"), userlib.RandomBytes(16), data)
	file.Data = append([][]byte(nil), encdata)
	fmac, _ := userlib.HMACEval(bHashKDF(fileinfo.FKey, "mac"), encdata)
	file.Mac = append([][]byte(nil), fmac)
	mfile, _ := json.Marshal(file)
	userlib.DatastoreSet(fileinfo.Fuuid, mfile)

	wrappermap, err := userdata.mapRevoke(filename, target_username)
	if err != nil {
		return err
	}

	var recfileinfo FileInfo
	for _, recwrapper := range wrappermap {
		encminfo, ok := userlib.DatastoreGet(recwrapper.Infouuid)
		if !ok {
			return errors.New(strings.ToTitle("File Info not found!"))
		}
		minfo := userlib.SymDec(recwrapper.InfoKey, encminfo)
		err = json.Unmarshal(minfo, &recfileinfo)
		if err != nil {
			return err
		}
		recfileinfo.FKey = fileinfo.FKey
		minfo, _ = json.Marshal(recfileinfo)
		encinfo = userlib.SymEnc(recwrapper.InfoKey, userlib.RandomBytes(16), minfo)
		userlib.DatastoreSet(recwrapper.Infouuid, encinfo)
	}
	return nil
}
