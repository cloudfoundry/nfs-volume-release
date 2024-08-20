package nfsv3driver_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/goshims/ldapshim/ldap_fake"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/nfsv3driver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/ldap.v2"
)

var _ = Describe("IdResolverTest", func() {
	var ldapFake *ldap_fake.FakeLdap
	var ldapConnectionFake *ldap_fake.FakeLdapConnection
	var ldapIdResolver nfsv3driver.IdResolver
	var env dockerdriver.Env
	var uid string
	var gid string
	var err error
	var ldapCACert string
	var ldapTimeout time.Duration
	var user string

	BeforeEach(func() {
		logger := lagertest.NewTestLogger("nfs-mounter")
		testContext := context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, testContext)

		user = "user"
	})

	JustBeforeEach(func() {
		ldapIdResolver = nfsv3driver.NewLdapIdResolver(
			"svcuser",
			"svcpw",
			"host",
			111,
			"tcp",
			"cn=Users,dc=test,dc=com",
			ldapCACert,
			ldapFake,
			ldapTimeout,
		)
		uid, gid, err = ldapIdResolver.Resolve(env, user, "pw")
	})

	Context("when the connection is successful", func() {
		BeforeEach(func() {
			ldapFake = &ldap_fake.FakeLdap{}
			ldapConnectionFake = &ldap_fake.FakeLdapConnection{}
			ldapFake.DialReturns(ldapConnectionFake, nil)
			ldapCACert = ""
			ldapTimeout = 120 * time.Second
			ldapConnectionFake.SearchReturns(&ldap.SearchResult{}, nil)
		})

		Context("when CA cert is provided", func() {
			BeforeEach(func() {
				ldapCACert = `-----BEGIN CERTIFICATE-----
MIIDGTCCAgGgAwIBAgIRAIlVvSGFPY1EvNayuTpPAScwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xODA1MzExNzU5MTBaFw0xOTA1MzExNzU5
MTBaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCSf8J68FYrRuE8+NumcleeI10+O5QGibQ3+axX79eFS3RGcQKn5UOr
OFE/RM/ghc7sUD8urLhlA2QAua+0dZEr+QtNswDxLfWljw08azR4xkPnBejdwYKU
jHHU9UoJrxEgWqNFwTWWCyHYERUK/RFSrSUJaZLv1fRa9C+wbkD2Wd+aesPU6TZr
5f6DT1UdL5umykwVoKy9ymA1CUi3iRSPuIxF0iuwwNtgtS0Dswi9+gqICOYp+lGJ
RM2zRZFas8clubvkIRYlO2YG8hb181uxW9nLAfUfJjjtDt7lp5z/eZqliFwzrl0i
DG8xWUppHV9654hGRDOL2ow3u8kwNv9/AgMBAAGjajBoMA4GA1UdDwEB/wQEAwIC
pDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MDAGA1UdEQQp
MCeCJW5mc3Rlc3RsZGFwc2VydmVyLnNlcnZpY2UuY2YuaW50ZXJuYWwwDQYJKoZI
hvcNAQELBQADggEBAB/4KT3+G5YqrnCCF+GmYlxZO9ScRA6yPBtwXTQe7WH8Yfz2
bnUs4jKhK2wh3+RSTsBwV9afF+xm/uVrD9iZveixC1E3NqJwlchHc2bv9NCvC8OY
VShIx+8Joqpud6VIrzclhus2lo9Dvn55at3Z/5SYDf07fDmSJ5pZuLUVryiJk9AT
G0GELNbBftMakAJaH6eqGvcNbDRMeFqq7VyjthQJRPWSaWKA6TsfzgiO9lwx1wd1
1ZtN1nl1NexFqcan26vg0f1SwLM9r9mVXrKII/T60RXKvtcAkMS3XfaebG3ulout
z6sbK6WkL0AwPEcI/HzUOrsAUBtyY8cfy6yVcuQ=
-----END CERTIFICATE-----`
				ldapFake.DialTLSReturns(ldapConnectionFake, nil)
			})

			It("connects via TLS", func() {
				Expect(ldapFake.DialTLSCallCount()).To(Equal(1))
				protocol, addr, config := ldapFake.DialTLSArgsForCall(0)
				Expect(protocol).To(Equal("tcp"))
				Expect(addr).To(Equal("host:111"))
				Expect(config.ServerName).To(Equal("host"))
				Expect(config.RootCAs.Subjects()).To(HaveLen(1))                             //lint:ignore SA1019 "not systemcert"
				Expect(string(config.RootCAs.Subjects()[0])).To(ContainSubstring("Acme Co")) //lint:ignore SA1019 "not systemcert"
			})
		})

		Context("when CA cert is not provided", func() {
			BeforeEach(func() {
				ldapCACert = ""
			})

			It("connects without TLS", func() {
				Expect(ldapFake.DialCallCount()).To(Equal(1))
				protocol, addr := ldapFake.DialArgsForCall(0)
				Expect(protocol).To(Equal("tcp"))
				Expect(addr).To(Equal("host:111"))
			})
		})

		Context("when search returns successfully", func() {
			BeforeEach(func() {
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{
						{Name: "uidNumber", Values: []string{"100"}},
						{Name: "gidNumber", Values: []string{"100"}},
					},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("should build a valid ldap search request", func() {
				baseDN, scope, derefAliases, sizeLimit, timeLimit, typesOnly, filter, attributes, controls := ldapFake.NewSearchRequestArgsForCall(0)
				Expect(baseDN).To(Equal("cn=Users,dc=test,dc=com"))
				Expect(scope).To(Equal(2))
				Expect(derefAliases).To(Equal(0))
				Expect(sizeLimit).To(Equal(0))
				Expect(timeLimit).To(Equal(0))
				Expect(typesOnly).To(BeFalse())
				Expect(filter).To(Equal("(&(objectClass=User)(cn=user))"))
				Expect(attributes).To(ConsistOf("dn", "uidNumber", "gidNumber"))
				Expect(controls).To(BeNil())
			})

			It("set timeout for connection", func() {
				Expect(ldapConnectionFake.SetTimeoutCallCount()).To(Equal(1))
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the expected UID and GID", func() {
				Expect(uid).To(Equal("100"))
				Expect(gid).To(Equal("100"))
			})

			Context("when the credentials are not good", func() {
				BeforeEach(func() {
					ldapConnectionFake.BindStub = func(u, p string) error {
						if u == "svcuser" {
							return nil
						} else {
							return errors.New("badness")
						}
					}
				})
				It("should find the user and then fail", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(BeAssignableToTypeOf(dockerdriver.SafeError{}))
					Expect(err.Error()).To(ContainSubstring("badness"))
					Expect(ldapConnectionFake.SearchCallCount()).To(Equal(1))
					Expect(uid).To(BeEmpty())
				})
			})
		})

		Context("when the search uses an invalid username", func() {
			BeforeEach(func() {
				user = "*"
			})

			It("should continue to search for the username", func() {
				_, _, _, _, _, _, req, _, _ := ldapFake.NewSearchRequestArgsForCall(0)
				Expect(req).To(Equal("(&(objectClass=User)(cn=\\2a))"))
			})
		})

		Context("when search does not return GID", func() {
			BeforeEach(func() {
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{
						{Name: "uidNumber", Values: []string{"100"}},
					},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("sets GID same as UID", func() {
				Expect(uid).To(Equal("100"))
				Expect(gid).To(Equal("100"))
			})
		})

		Context("when search returns empty GID", func() {
			BeforeEach(func() {
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{
						{Name: "uidNumber", Values: []string{"100"}},
						{Name: "gidNumber", Values: []string{""}},
					},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("sets GID same as UID", func() {
				Expect(uid).To(Equal("100"))
				Expect(gid).To(Equal("100"))
			})
		})

		Context("when the search returns empty", func() {
			BeforeEach(func() {
				result := &ldap.SearchResult{Entries: []*ldap.Entry{}}
				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("reports an error for the missing user", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("User does not exist"))
				Expect(err).To(BeAssignableToTypeOf(dockerdriver.SafeError{}))
			})
		})

		Context("when the search returns multiple results", func() {
			BeforeEach(func() {
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{
						{Name: "uidNumber", Values: []string{"100"}},
						{Name: "gidNumber", Values: []string{"100"}},
					},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry, entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("reports an error for the ambiguous search", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Ambiguous search--too many results"))
				Expect(err).To(BeAssignableToTypeOf(dockerdriver.SafeError{}))
			})
		})
	})

	Context("LDAP Server is unreachable", func() {
		BeforeEach(func() {
			ldapFake = &ldap_fake.FakeLdap{}
			ldapConnectionFake = &ldap_fake.FakeLdapConnection{}
			ldapFake.DialReturns(ldapConnectionFake, errors.New("unable to reach ldap server"))

			ldapCACert = ""
			ldapTimeout = 120 * time.Second
			ldapConnectionFake.SearchReturns(&ldap.SearchResult{}, nil)
		})

		It("Should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("LDAP server could not be reached, please contact your system administrator"))
			Expect(err).To(BeAssignableToTypeOf(dockerdriver.SafeError{}))
		})
	})
})
