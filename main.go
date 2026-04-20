package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/naivary/dcv-virtual-session-ldap/dcv"
)

const LDAPAttrLogonName = "sAMAccountName"

type flagOpts struct {
	period   time.Duration
	ldapURL  string
	username string
	password string
	groupDN  string
	baseDN   string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	opts := flagOpts{}
	flag.DurationVar(&opts.period, "period", 5*time.Minute, "period at which the LDAP users are re-read.")
	flag.StringVar(&opts.ldapURL, "url", "", "LDAP URL. Supported schemas are ldap://, ldaps://, ldapi://, cldap://")
	flag.StringVar(&opts.username, "user", "", "Username to use for LDAP binding.")
	flag.StringVar(&opts.password, "password", "", "Password of the username.")
	flag.StringVar(&opts.groupDN, "gdn", "", "Group DN users have to be a member of to create a virtual session for.")
	flag.StringVar(&opts.baseDN, "bdn", "", "Base DN of the LDAP.")
	flag.Parse()

	conn, err := ldap.DialURL(opts.ldapURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Bind(opts.username, opts.password)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(opts.period)
	defer ticker.Stop()
	for range ticker.C {
		users, err := listDCVMembers(conn, opts.baseDN, opts.groupDN)
		if err != nil {
			return err
		}
		for _, user := range users {
			err = dcv.CreateVirtualSessionFromUsername(user)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

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
		userLogonName := e.GetAttributeValue(LDAPAttrLogonName)
		if userLogonName == "" {
			return nil, fmt.Errorf("undefined logon name: %s", e.DN)
		}
		users = append(users, userLogonName)
	}
	return users, nil
}
