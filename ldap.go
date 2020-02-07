package main

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

func ldapMail(cfg config, users []string) []string {
	var filter string
	var mail []string

	if len(users) == 0 {
		return mail
	}

	for _, item := range users {
		filter = fmt.Sprintf("%v(sAMAccountName=%v)", filter, item)
	}

	conn, err := ldapConnect(cfg)

	if err != nil {
		fmt.Printf("Failed to connect. %s", err)
		return mail
	}

	defer conn.Close()

	filter = fmt.Sprintf("(&(objectClass=user)(objectCategory=person)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(|%v))", filter)

	if err, mail = ldapList(conn, cfg.LBase, filter); err != nil {
		fmt.Printf("%v", err)
		return mail
	}

	return mail

}

func ldapConnect(cfg config) (*ldap.Conn, error) {
	conn, err := ldap.Dial("tcp", cfg.LHost)

	if err != nil {
		return nil, fmt.Errorf("Failed to connect. %s", err)
	}

	if err := conn.Bind(cfg.LUser, cfg.LPass); err != nil {
		return nil, fmt.Errorf("Failed to bind. %s", err)
	}

	return conn, nil
}

func ldapList(conn *ldap.Conn, base string, filter string) (error, []string) {
	var mail []string

	result, err := conn.Search(ldap.NewSearchRequest(
		base,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{"mail"},
		nil,
	))

	if err != nil {
		return fmt.Errorf("Failed to search users. %s", err), mail
	}

	for _, entry := range result.Entries {
		mail = append(mail, entry.GetAttributeValue("mail"))
	}

	return nil, mail
}
