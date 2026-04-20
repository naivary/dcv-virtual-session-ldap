package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/naivary/dcv-virtual-session-ldap/dcv"
)

const (
	LDAPAttrLogonName     = "sAMAccountName"
	LDAPAttrPrincipalName = "userPrincipalName"
)

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
		fmt.Fprintf(os.Stderr, "err: %s\n", err)
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
	slog.Info("Successfully started DCV Virtual Session Manager!", "period", opts.period)
	for range ticker.C {
		users, err := listDCVMembers(conn, opts.baseDN, opts.groupDN)
		if err != nil {
			return err
		}
		for _, email := range users {
			name, _, _ := strings.Cut(email, "@")
			vs := dcv.VirtualSession{
				ID:    name,
				Owner: email,
			}
			err = dcv.CreateVirtualSession(&vs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
