require 'rspec'
require 'bosh/template/test'

describe 'nfsv3driver job' do
  let(:release) {Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..'))}
  let(:job) {release.job('nfsv3driver')}

  describe 'start.sh' do
    let(:template) {job.template('bin/start.sh')}
    mapfs_link = [
        Bosh::Template::Test::Link.new(
            name: 'mapfs',
            properties: {
                'path' => '/some/path',
            }
        )
    ]

    context 'when configured with a ca cert' do
      let(:manifest_properties) do
        {
            "nfsv3driver" => {
                "tls" => {
                    "ca_cert" => "some-ca-cert",
                },
            }
        }
      end

      it 'successfully renders the script' do
        tpl_output = template.render(manifest_properties, consumes: mapfs_link)

        expect(tpl_output).to include("--caFile=\"${CLIENT_CERTS_DIR}/ca.crt\"")
      end
    end

      context 'when configured with ldap enabled' do
        let(:manifest_properties) do
          {
            "nfsv3driver" => {
              "ldap_svc_user" => "service-user",
              "ldap_svc_password" => "service-password",
              "ldap_host" => "some-host",
              "ldap_port" => 1234,
              "ldap_proto" => "udp",
              "ldap_user_fqdn" => "cn=Users,dc=corp,dc=test,dc=com",
              "ldap_ca_cert" => "some-ca-cert",
            }
          }
        end

        it 'sets the allowedOptions flag correctly' do
          tpl_output = template.render(manifest_properties, consumes: mapfs_link)

          expect(tpl_output).to include("export LDAP_SVC_USER='service-user'")
          expect(tpl_output).to include("export LDAP_SVC_PASS='service-password'")
          expect(tpl_output).to include("export LDAP_HOST=\"some-host\"")
          expect(tpl_output).to include("export LDAP_PORT=\"1234\"")
          expect(tpl_output).to include("export LDAP_PROTO=\"udp\"")
          expect(tpl_output).to include("export LDAP_USER_FQDN=\"cn=Users,dc=corp,dc=test,dc=com\"")
          expect(tpl_output).to include("export LDAP_CA_CERT=\"some-ca-cert\"")
        end
      end

      context 'when ldap properties contain bash special characters' do
        let(:manifest_properties) do
          {
            "nfsv3driver" => {
              "ldap_svc_user" => "Patrick O'Malley",
              "ldap_svc_password" => "!que&pasa!${xxx}$?",
              "ldap_host" => "some-host",
              "ldap_port" => 1234,
              "ldap_proto" => "udp",
              "ldap_user_fqdn" => "cn=Users,dc=corp,dc=test,dc=com",
              "ldap_ca_cert" => "some-ca-cert",
            }
          }
        end

        it 'escapes the properties correctly' do
          tpl_output = template.render(manifest_properties, consumes: mapfs_link)

          expect(tpl_output).to include("export LDAP_SVC_USER='Patrick O'\"'\"'Malley'")
          expect(tpl_output).to include("export LDAP_SVC_PASS='!que&pasa!${xxx}$?'")
        end
      end
    context 'when configured with ldap with a null ca cert' do
      let(:manifest_properties) do
        {
            "nfsv3driver" => {
                "ldap_svc_user" => "service-user",
                "ldap_svc_password" => "service-password",
                "ldap_host" => "some-host",
                "ldap_port" => 1234,
                "ldap_proto" => "udp",
                "ldap_user_fqdn" => "cn=Users,dc=corp,dc=test,dc=com",
                "ldap_ca_cert" => nil,
            }
        }
      end

      it 'renders LDAP_CA_CERT with an empty string' do
        tpl_output = template.render(manifest_properties, consumes: mapfs_link)

        expect(tpl_output).to include("export LDAP_CA_CERT=\"\"")
      end
    end

  end
end
