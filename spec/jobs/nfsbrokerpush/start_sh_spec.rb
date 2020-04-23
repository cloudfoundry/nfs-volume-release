require 'rspec'
require 'bosh/template/test'

describe 'nfsbrokerpush job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('nfsbrokerpush') }

  describe 'start.sh' do
    let(:template) { job.template('start.sh') }
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
    ]}

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
          }
        }
      end

      it 'successfully renders the script' do
        tpl_output = template.render(manifest_properties, consumes: credhub_link)

        expect(tpl_output).to include("bin/nfsbroker --listenAddr=\"0.0.0.0:$PORT\"")
        expect(tpl_output).to include("--credhubURL=\"https://some-credhub-url:4321\"")
        expect(tpl_output).to include("--uaaClientID=\"client-id\"")
        expect(tpl_output).to include("--uaaClientSecret=\"client-secret\"")
        expect(tpl_output).to include("--servicesConfig=\"./services.json\"")
        expect(tpl_output).to include("--logLevel=\"some-log-level\"")
        expect(tpl_output).to include("--timeFormat=\"some-log-time-format\"")
        expect(tpl_output).to include("--allowedOptions=\"source,uid,gid,auto_cache,readonly,version,mount,cache\"")
      end
    end

    context 'when configured with all required credhub properties' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
            "credhub" => {
              "uaa_client_id" => "some-uaa-client-id",
              "uaa_client_secret" => "some-uaa-client-secret",
            }
          }
        }
      end

      it 'includes the credhub flags in the script' do
        tpl_output = template.render(manifest_properties, consumes: credhub_link)

        expect(tpl_output).to include("--credhubURL=\"https://some-credhub-url:4321\"")
        expect(tpl_output).to include("--uaaClientID=\"some-uaa-client-id\"")
        expect(tpl_output).to include("--uaaClientSecret=\"some-uaa-client-secret\"")
        expect(tpl_output).to include("--storeID=\"nfsbroker\"")
      end

      context 'configured with credhub set to zero instances' do
        let(:credhub_link) { [
          Bosh::Template::Test::Link.new(
            name: 'credhub',
            instances: []
          )
        ]}

        it 'a meaningful error message is returned' do
          expect{template.render(manifest_properties, consumes: credhub_link)}.to raise_error('credhub is required. Zero instances found.')
        end
      end

      context 'configured with no credhub link' do
        it 'a meaningful error message is returned' do
          expect{template.render(manifest_properties)}.to raise_error("Can't find link 'credhub'")
        end
      end
    end

    context 'when configured with no credhub properties' do
      let(:manifest_properties) do
        {}
      end

      it 'fails with a meaningful error message' do
        expect { template.render(manifest_properties, consumes: credhub_link) }.to raise_error('missing credhub uaa properties')
      end
    end

    context 'when configured with ldap enabled' do
      let(:manifest_properties) do
        {
          "nfsbrokerpush" => {
           "credhub" => {
              "uaa_client_id" => "client-id",
              "uaa_client_secret" => "client-secret",
            },
            "ldap_enabled" => true,
          }
        }
      end

      it 'sets the allowedOptions flag correctly' do
        tpl_output = template.render(manifest_properties, consumes: credhub_link)

        expect(tpl_output).to include("--allowedOptions=\"source,auto_cache,username,password,readonly,version,mount,cache\"")
      end
    end
  end
end
