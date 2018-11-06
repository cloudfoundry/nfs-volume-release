require 'rspec'
require 'bosh/template/test'

describe 'nfsbrokerpush job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('nfsbrokerpush') }

  describe 'start.sh' do
    let(:template) { job.template('start.sh') }

    context 'when configured with all database properties' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "host" => "some-db-host",
              "port" => "some-db-port",
              "name" => "some-db-name",
              "driver" => "some-db-driver",
              "ca_cert" => "some-ca-cert",
            },
            "store_id" => "some-store-id",
            "log_level" => "some-log-level",
            "log_time_format" => "some-log-time-format",
          }
        }
      end

      it 'successfully renders the script with the db flags' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("bin/nfsbroker --listenAddr=\"0.0.0.0:$PORT\"")
        expect(tpl_output).to include("--servicesConfig=\"./services.json\"")
        expect(tpl_output).to include("--dbDriver=\"some-db-driver\"")
        expect(tpl_output).to include("--dbHostname=\"some-db-host\"")
        expect(tpl_output).to include("--dbPort=\"some-db-port\"")
        expect(tpl_output).to include("--dbName=\"some-db-name\"")
        expect(tpl_output).to include("--credhubURL=\"\"")
        expect(tpl_output).to include("--uaaClientID=\"\"")
        expect(tpl_output).to include("--uaaClientSecret=\"\"")
        expect(tpl_output).to include("--storeID=\"some-store-id\"")
        expect(tpl_output).to include("--logLevel=\"some-log-level\"")
        expect(tpl_output).to include("--timeFormat=\"some-log-time-format\"")
        expect(tpl_output).to include("--allowedOptions=\"uid,gid,auto_cache,version\"")
        expect(tpl_output).to include("--dbCACertPath=\"./db_ca.crt\"")
      end
    end

    context 'when configured with the database host via a link' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "port" => "some-db-port",
              "name" => "some-db-name",
              "driver" => "some-db-driver",
              "ca_cert" => "some-ca-cert",
            }
          }
        }
      end

      it 'sets the dbHostname flag from the link address' do
        links = [
          Bosh::Template::Test::Link.new(
            name: 'database',
            instances: [Bosh::Template::Test::LinkInstance.new(address: 'some-db-host-from-link')],
            properties: {},
          )
        ]

        tpl_output = template.render(manifest_properties, consumes: links)

        expect(tpl_output).to include("--dbHostname=\"some-db-host-from-link\"")
      end
    end

    context 'when configured to skip database hostname validation' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "host" => "some-db-host",
              "port" => "some-db-port",
              "name" => "some-db-name",
              "driver" => "some-db-driver",
              "ca_cert" => "some-ca-cert",
              "skip_hostname_validation" => true,
            }
          }
        }
      end

      it 'includes the dbSkipHostnameValidation flag in the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("--dbSkipHostnameValidation")
      end
    end

    context 'when configured with all credhub properties' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "host" => "some-db-host",
            },
            "credhub" => {
              "url" => "some-credhub-url",
              "uaa_client_id" => "some-uaa-client-id",
              "uaa_client_secret" => "some-uaa-client-secret",
            }
          }
        }
      end

      it 'includes the credhub flags' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("--credhubURL=\"some-credhub-url\"")
        expect(tpl_output).to include("--uaaClientID=\"some-uaa-client-id\"")
        expect(tpl_output).to include("--uaaClientSecret=\"some-uaa-client-secret\"")
      end
    end

    context 'when configured with ldap enabled' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "host" => "some-db-host",
            },
            "ldap_enabled" => true,
          }
        }
      end

      it 'sets the allowedOptions flag correctly' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("--allowedOptions=\"auto_cache,username,password,version\"")
      end
    end

    context 'when configured with ldap test mode enabled' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "host" => "some-db-host",
            },
            "ldap_test_mode" => true,
          }
        }
      end

      it 'sets the allowedOptions flag correctly' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("--allowedOptions=\"auto_cache,uid,gid,username,password,version\"")
      end
    end
  end
end
