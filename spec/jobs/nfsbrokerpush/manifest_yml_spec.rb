require 'rspec'
require 'bosh/template/test'

describe 'nfsbrokerpush job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('nfsbrokerpush') }

  describe 'manifest.yml' do
    let(:template) { job.template('manifest.yml') }
    let(:credhub_link) { [
      Bosh::Template::Test::Link.new(
        name: 'credhub',
        instances: [Bosh::Template::Test::LinkInstance.new(address: 'credhub.service.cf.internal')],
        properties: {
          'credhub' => {
            'internal_url' => 'some-credhub-url',
            'port' => 4321,
            'ca_certificate' => 'some-certificate',
          }
        }
      )
    ] }

    context 'when fully configured with all required credhub properties' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "credhub" => {
              "uaa_client_id" => "client-id",
              "uaa_client_secret" => "client-secret",
            },
            "store_id" => "some-store-id",
            "log_level" => "some-log-level",
            "log_time_format" => "some-log-time-format",
            "app_name" => "super-cool-app",
            "app_domain" => "cf-domain.test",
            "username" => "jane-doe",
            "password" => "fake-secret",
          }
        }
      end

      it 'successfully renders the manifest' do
        tpl_output = template.render(manifest_properties, consumes: credhub_link)

        expect(tpl_output).to include("---")
        expect(tpl_output).to include('- name: "super-cool-app"')
        expect(tpl_output).to include('  buildpacks:')
        expect(tpl_output).to include('  - binary_buildpack')
        expect(tpl_output).to include('  routes:')
        expect(tpl_output).to include('  - route: "super-cool-app.cf-domain.test"')
        expect(tpl_output).to include('  memory: "256M"')
        expect(tpl_output).to include('  env:')
        expect(tpl_output).to include('    USERNAME: "jane-doe"')
        expect(tpl_output).to include('    PASSWORD: "fake-secret"')
        expect(tpl_output).to include('    UAA_CLIENT_ID: "client-id"')
        expect(tpl_output).to include('    UAA_CLIENT_SECRET: "client-secret"')
      end
    end

    context 'when UAA properties are missing' do
      let(:manifest_properties) {
        {
          "nfsbrokerpush" => {
            "store_id" => "some-store-id",
            "log_level" => "some-log-level",
            "log_time_format" => "some-log-time-format",
            "app_name" => "super-cool-app",
            "app_domain" => "cf-domain.test",
            "username" => "jane-doe",
            "password" => "fake-secret",
          }
        }
      }

      it 'should present a meaningful message' do
        expect {template.render(manifest_properties, consumes: credhub_link) }.to raise_error("missing credhub UAA credentials")
      end
    end
  end
end