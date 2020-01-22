package main

import (
	"fmt"

	"gopkg.in/ldap.v3"
)

func ldapMail(cfg config, user string) []string {
	conn, err := ldapConnect(cfg)

	var mail []string

	if err != nil {
		fmt.Printf("Failed to connect. %s", err)
		return mail
	}

	defer conn.Close()

	filter := fmt.Sprintf("(sAMAccountName=%v)", user)
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
