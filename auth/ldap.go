package auth

import (
	b64 "encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/go-ldap/ldap/v3"
	dbase "github.com/javitab/go-web/database"
)

// UserLogin struct to represent user credentials
type UserLDAPLogin struct {
	Username string
	Password string
}

// LdapBindCredentials struct to represent credentials to bind with LDAP server for lookup
type LdapBindCredentials struct {
	Username string
	Password string
}

// Decode LdapBindCredentials

func GetLdapBindCredentials() LdapBindCredentials {
	encodedCreds := os.Getenv("LDAP_BIND_CREDENTIALS")
	decodedCreds, _ := b64.StdEncoding.DecodeString(encodedCreds)
	decodedCredsString := string([]byte(decodedCreds))
	ldapBindParts := strings.Split(decodedCredsString, ":")

	// Create object
	var LdapBindCredentials LdapBindCredentials
	LdapBindCredentials.Username = ldapBindParts[0]
	LdapBindCredentials.Password = ldapBindParts[1]

	// Return object
	return LdapBindCredentials
}

func GetLdapConnection() (*ldap.Conn, error) {
	conn, err := ldap.DialURL(os.Getenv("LDAP_ADDRESS"))
	if err != nil {
		dbase.LogServerError("GetLdapConnection:ldap.DialURL", err, "Error loading connection: "+os.Getenv("LDAP_ADDRESS"))
		return nil, err
	}
	creds := GetLdapBindCredentials()
	if err := conn.Bind(creds.Username, creds.Password); err != nil {
		dbase.LogServerError("GetLdapConnection:ldap.Bind", err, "Error binding LDAP connection: "+creds.Username)
		return nil, err
	}
	return conn, nil
}

func LDAPAuth(creds LoginUserInput) (bool, error) {
	conn, err := GetLdapConnection()
	if err != nil {
		dbase.LogServerError("LDAPAuth:GetLdapConnection", err, "Failed to get connection to perform LDAPAuth request: "+creds.Username)
		return false, err
	}
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAP_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(anr=%s)", creds.Username),
		[]string{"dn"},
		nil,
	)
	searchResp, err := conn.Search(searchRequest)
	if err != nil {
		dbase.LogServerError("LDAPAuth:searchRequest:RequestFailed", err, fmt.Sprintf("LDAP search failed for user %s, error details: %v", creds.Username, err))
		return false, err
	}
	if len(searchResp.Entries) == 0 {
		err = fmt.Errorf("user: %s not found", creds.Username)
		dbase.LogServerError("LDAPAuth:searchRequest:NotFound", err, "Login Error Logged")
		return false, err
	}

	userDN := searchResp.Entries[0].DN

	err = conn.Bind(userDN, creds.Password)
	if err != nil {
		err = fmt.Errorf("authentication failed")
		dbase.LogServerError("LDAPAuth:ldapAuthBind:authFailed", err, "")
		return false, err
	}
	return true, nil
}

func LDAPGroups(username string) (groups []string) {
	conn, _ := GetLdapConnection()

	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAP_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(sAMAccountName=%s)", username),
		[]string{"memberOf"},
		nil,
	)

	searchResult, err := conn.Search(searchRequest)
	if err != nil {
		dbase.LogServerError("LDAPGroups:ldapSearchRequest:error", err, "")
		return nil
	}
	if len(searchResult.Entries) > 0 {
		user := searchResult.Entries[0]
		for _, entry := range user.GetAttributeValues("memberOf") {
			groupCN := strings.Split(entry, ",")[0]
			groupCN = strings.Replace(groupCN, "CN=", "", 1)
			groups = append(groups, groupCN)
		}
	}
	return groups
}

type LDAPUserInfo struct {
	Groups    []string
	Email     string
	LastName  string
	FirstName string
}

func LDAPGetUserInfo(username string) (LDAPUserInfo, error) {
	UserInfo := LDAPUserInfo{}
	conn, _ := GetLdapConnection()
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAP_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(sAMAccountName=%s)", username),
		[]string{"mail", "givenName", "sn"},
		nil,
	)

	searchResult, err := conn.Search(searchRequest)
	if err != nil {
		dbase.LogServerError("LDAPGetUserInfo:ldapSearchRequest:error", err, "")
		return UserInfo, err
	}
	if len(searchResult.Entries) != 1 {
		lookup_err := fmt.Errorf(("user not found in LDAP"))
		dbase.LogServerError("LDAPGetUserInfo:UserNotFound", lookup_err, fmt.Sprintf("User: %v not found in LDAP", username))
		return UserInfo, lookup_err
	}
	user := searchResult.Entries[0]
	UserInfo.Groups = LDAPGroups(username)
	UserInfo.Email = user.GetAttributeValue("mail")
	UserInfo.LastName = user.GetAttributeValue("sn")
	UserInfo.FirstName = user.GetAttributeValue("givenName")
	return UserInfo, nil
}

func LDAPUserExists(username string) bool {
	_, err := LDAPGetUserInfo(username)
	return err == nil
}

func LDAPEvalGroups(u UserInfo) {

	// Function to get the difference between two slices
	getDiff := func(arr1, arr2 []int) []int {
		diff := []int{}
		m := make(map[int]bool)

		// Add all elements of arr2 to the map
		for _, num := range arr2 {
			m[num] = true
		}

		// Check for elements in arr1 that are not in arr2
		for _, num := range arr1 {
			if !m[num] {
				diff = append(diff, num)
			}
		}

		return diff
	}

	db := dbase.GetDBConn()
	var LDAPGroupIDs []int
	db.Raw("SELECT id from public.groups WHERE ldap_group in ?", u.LDAPGroups).Scan(&LDAPGroupIDs)

	// Get all groups that user is in
	var UserGroups []int
	db.Raw("select group_id from user_groups where user_id = ?", u.DB.ID).Scan(&UserGroups)

	addGroups := getDiff(LDAPGroupIDs, UserGroups)
	remGroups := getDiff(UserGroups, LDAPGroupIDs)

	// Remove users in RemGroups
	for _, groupID := range remGroups {
		db.Exec("DELETE from user_groups where user_id = ? AND group_id = ?", u.DB.ID, groupID)
		dbase.LogServerEvent("LDAPEvalGroups:LDAPRemoveGroup", fmt.Sprintf("LDAP mandated remove user %v from group id %v", u.DB.Username, groupID), "LDAP")
	}

	// Add user to groups in AddGroups
	for _, groupID := range addGroups {
		db.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", u.DB.ID, groupID)
		dbase.LogServerEvent("LDAPEvalGroups:LDAPAddGroup", fmt.Sprintf("LDAP mandated add user %v to group id %v", u.DB.Username, groupID), "LDAP")
	}
}
