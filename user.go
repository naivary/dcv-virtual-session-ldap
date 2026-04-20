package main

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

func listDCVMembers(conn *ldap.Conn, baseDN, groupDN string) ([]string, error) {
	const (
		sizeLimit = 0
		timeLimit = 0
		typesOnly = false
	)
	filter := fmt.Sprintf("(&(objectClass=user)(memberOf=%s,%s))", groupDN, baseDN)
	attrs := []string{}
	req := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		sizeLimit, timeLimit,
		typesOnly,
		filter,
		attrs,
		nil,
	)

	sr, err := conn.Search(req)
	if err != nil {
		return nil, err
	}
	users := make([]string, 0, len(sr.Entries))
	for _, e := range sr.Entries {
		principalName := e.GetAttributeValue(LDAPAttrPrincipalName)
		if principalName == "" {
			return nil, fmt.Errorf("undefined principal name: %s", e.DN)
		}
		users = append(users, principalName)
	}
	return users, nil
}
