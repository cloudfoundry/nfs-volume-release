require 'rspec'
require 'bosh/template/test'

describe 'nfsbrokerpush job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('nfsbrokerpush') }

  describe 'start.sh' do
    let(:template) { job.template('start.sh') }

    context 'when fully configured with all required database properties' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "host" => "some-db-host",
              "port" => "some-db-port",
              "name" => "some-db-name",
              "driver" => "some-db-driver",
            },
            "store_id" => "some-store-id",
            "log_level" => "some-log-level",
            "log_time_format" => "some-log-time-format",
          }
        }
      end

      it 'successfully renders the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("bin/nfsbroker --listenAddr=\"0.0.0.0:$PORT\"")
        expect(tpl_output).to include("--servicesConfig=\"./services.json\"")
        expect(tpl_output).to include("--storeID=\"some-store-id\"")
        expect(tpl_output).to include("--logLevel=\"some-log-level\"")
        expect(tpl_output).to include("--timeFormat=\"some-log-time-format\"")
        expect(tpl_output).to include("--allowedOptions=\"uid,gid,auto_cache,version\"")
      end

      it 'includes the db flags in the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("--dbDriver=\"some-db-driver\"")
        expect(tpl_output).to include("--dbHostname=\"some-db-host\"")
        expect(tpl_output).to include("--dbPort=\"some-db-port\"")
        expect(tpl_output).to include("--dbName=\"some-db-name\"")
      end

      it 'omits the dbSkipHostnameValidation flag from the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).not_to include("--dbSkipHostnameValidation")
      end

      it 'omits the dbCACertPath flag from the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).not_to include("--dbCACertPath=")
      end

      it 'omits the credhub flags from the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).not_to include("--credhubURL=")
        expect(tpl_output).not_to include("--uaaClientID=")
        expect(tpl_output).not_to include("--uaaClientSecret=")
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

        expect(tpl_output).to include("--dbDriver=\"some-db-driver\"")
        expect(tpl_output).to include("--dbHostname=\"some-db-host-from-link\"")
        expect(tpl_output).to include("--dbPort=\"some-db-port\"")
        expect(tpl_output).to include("--dbName=\"some-db-name\"")
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

    context 'when configured with a database CA cert' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "host" => "some-db-host",
              "port" => "some-db-port",
              "name" => "some-db-name",
              "driver" => "some-db-driver",
              "ca_cert" => "some-ca-cert",
            }
          }
        }
      end

      it 'includes the dbCACertPath flag in the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("--dbCACertPath=\"./db_ca.crt\"")
      end
    end

    context 'when configured with all required credhub properties' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "credhub" => {
              "url" => "some-credhub-url",
              "uaa_client_id" => "some-uaa-client-id",
              "uaa_client_secret" => "some-uaa-client-secret",
            }
          }
        }
      end

      it 'includes the credhub flags in the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("--credhubURL=\"some-credhub-url\"")
        expect(tpl_output).to include("--uaaClientID=\"some-uaa-client-id\"")
        expect(tpl_output).to include("--uaaClientSecret=\"some-uaa-client-secret\"")
      end

      it 'omits the database flags from the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).not_to include("--dbDriver=")
        expect(tpl_output).not_to include("--dbHostname=")
        expect(tpl_output).not_to include("--dbPort=")
        expect(tpl_output).not_to include("--dbName=")
      end
    end

    context 'when configured with no credhub or database properties' do
      let(:manifest_properties) do
        {}
      end

      it 'omits the credhub flags from the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).not_to include("--credhubURL=")
        expect(tpl_output).not_to include("--uaaClientID=")
        expect(tpl_output).not_to include("--uaaClientSecret=")
      end

      it 'omits the database flags from the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).not_to include("--dbDriver=")
        expect(tpl_output).not_to include("--dbHostname=")
        expect(tpl_output).not_to include("--dbPort=")
        expect(tpl_output).not_to include("--dbName=")
      end
    end

    context 'when configured without a database host property or link' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "db" => {
              "port" => "some-db-port",
              "name" => "some-db-name",
              "driver" => "some-db-driver",
            }
          }
        }
      end

      it 'raises an error' do
        expect{template.render(manifest_properties)}.to raise_error('missing database host property or link')
      end
    end

    context 'when configured with ldap enabled' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
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
