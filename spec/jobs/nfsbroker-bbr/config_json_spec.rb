require 'rspec'
require 'json'
require 'bosh/template/test'

describe 'nfsbroker-bbr job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('nfsbroker-bbr') }

  describe 'config.json' do
    let(:template) { job.template('config/config.json') }

    context 'when configured with minimal properties' do
      let(:manifest_properties) do
        {
          "nfsbroker" => {
            "db_hostname" => "some-db-host",
            "db_username" => "some-db-user",
            "db_password" => "some-db-password",
            "db_port" => "some-db-port",
            "db_name" => "some-db-name",
            "db_driver" => "some-db-driver",
          }
        }
      end

      it 'generates successfully' do
        tpl_output = template.render(manifest_properties)

        config = JSON.parse(tpl_output)
        expect(config).to eq({
          "username" => "some-db-user",
          "password" => "some-db-password",
          "host" => "some-db-host",
          "port" => "some-db-port",
          "database" => "some-db-name",
          "adapter" => "some-db-driver",
        })
      end
    end

    context 'when a CA cert is specified' do
      let(:manifest_properties) do
        {
          "nfsbroker" => {
            "db_hostname" => "some-db-host",
            "db_username" => "some-db-user",
            "db_password" => "some-db-password",
            "db_port" => "some-db-port",
            "db_name" => "some-db-name",
            "db_driver" => "some-db-driver",
            "db_ca_cert" => "some-ca-cert",
          }
        }
      end

      it 'includes the CA cert' do
        tpl_output = template.render(manifest_properties)

        config = JSON.parse(tpl_output)
        expect(config).to eq({
          "username" => "some-db-user",
          "password" => "some-db-password",
          "host" => "some-db-host",
          "port" => "some-db-port",
          "database" => "some-db-name",
          "adapter" => "some-db-driver",
          "tls" => {
            "cert" => {
              "ca" => "some-ca-cert",
            }
          }
        })
      end
    end

    context 'when db_skip_hostname_validation is true' do
      let(:manifest_properties) do
        {
          "nfsbroker" => {
            "db_hostname" => "some-db-host",
            "db_username" => "some-db-user",
            "db_password" => "some-db-password",
            "db_port" => "some-db-port",
            "db_name" => "some-db-name",
            "db_driver" => "some-db-driver",
            "db_ca_cert" => "some-ca-cert",
            "db_skip_hostname_validation" => true,
          }
        }
      end

      it 'sets the flag to skip db hostname validation' do
        tpl_output = template.render(manifest_properties)

        config = JSON.parse(tpl_output)
        expect(config).to eq({
          "username" => "some-db-user",
          "password" => "some-db-password",
          "host" => "some-db-host",
          "port" => "some-db-port",
          "database" => "some-db-name",
          "adapter" => "some-db-driver",
          "tls" => {
            "cert" => {
              "ca" => "some-ca-cert",
            },
            "skip_host_verify" => true,
          }
        })
      end
    end

    context 'if the CA cert is missing' do
      let(:manifest_properties) do
        {
          "nfsbroker" => {
            "db_hostname" => "some-db-host",
            "db_username" => "some-db-user",
            "db_password" => "some-db-password",
            "db_port" => "some-db-port",
            "db_name" => "some-db-name",
            "db_driver" => "some-db-driver",
            "db_skip_hostname_validation" => true,
          }
        }
      end

      it 'ignores the flag to skip db hostname validation' do
        tpl_output = template.render(manifest_properties)

        config = JSON.parse(tpl_output)
        expect(config).to eq({
          "username" => "some-db-user",
          "password" => "some-db-password",
          "host" => "some-db-host",
          "port" => "some-db-port",
          "database" => "some-db-name",
          "adapter" => "some-db-driver",
        })
      end
    end
  end
end
