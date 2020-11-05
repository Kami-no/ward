package ldap

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"

	"github.com/Kami-no/ward/config"
)

type Service interface {
	Check(email string) bool
	ListMails(users []string) []string
}

type serviceImpl struct {
	cfg *config.Config
}

var _ Service = (*serviceImpl)(nil)

func NewLdapServiceImpl(cfg *config.Config) *serviceImpl {
	return &serviceImpl{cfg: cfg}
}

func (s *serviceImpl) Check(email string) bool {
	filter := fmt.Sprintf("(mail=%v)", email)

	mail := s.ldapRequest(filter)

	return len(mail) > 0
}

func (s *serviceImpl) ListMails(users []string) []string {
	var filter string

	if len(users) == 0 {
		return nil
	}

	for _, item := range users {
		filter = fmt.Sprintf("%v(sAMAccountName=%v)", filter, item)
	}

	mail := s.ldapRequest(filter)

	return mail
}

func (s *serviceImpl) ldapRequest(filter string) []string {
	var mail []string

	conn, err := s.ldapConnect()

	if err != nil {
		fmt.Printf("LDAP failed: %s", err)
		return nil
	}

	defer conn.Close()

	filter = fmt.Sprintf("(&(objectClass=user)(objectCategory=person)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(|%v))", filter)

	if mail, err = s.ldapList(conn, s.cfg.Endpoints.DC.Base, filter); err != nil {
		fmt.Printf("%v", err)
		return nil
	}

	return mail
}

func (s *serviceImpl) ldapConnect() (*ldap.Conn, error) {
	addr := fmt.Sprintf("%v:%v", s.cfg.Endpoints.DC.Host, s.cfg.Endpoints.DC.Port)
	conn, err := ldap.Dial("tcp", addr)

	if err != nil {
		return nil, fmt.Errorf("connection error: %s", err)
	}

	user := fmt.Sprintf("%v@%v", s.cfg.Credentials.User, s.cfg.Endpoints.DC.Domain)

	if err := conn.Bind(user, s.cfg.Credentials.Password); err != nil {
		return nil, fmt.Errorf("binding error: %s", err)
	}

	return conn, nil
}

func (s *serviceImpl) ldapList(conn *ldap.Conn, base string, filter string) ([]string, error) {
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
		return mail, fmt.Errorf("Failed to search users. %s", err)
	}

	for _, entry := range result.Entries {
		mail = append(mail, entry.GetAttributeValue("mail"))
	}

	return mail, nil
}
