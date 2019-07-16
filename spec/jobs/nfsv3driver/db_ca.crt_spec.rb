require 'rspec'
require 'bosh/template/test'

describe 'nfsv3driver job' do
  let(:release) {Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..'))}
  let(:job) {release.job('nfsv3driver')}

  describe 'ca.crt.erb' do
    let(:template) {job.template('config/certs/ca.crt')}

    context 'when fully configured with all required database properties' do
      let(:manifest_properties) do
        {
            "nfsv3driver" => {
                "tls" => {
                    "ca_cert" => "some-db-ca-cert",
                },
            }
        }
      end

      it 'successfully renders the script' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to include("some-db-ca-cert")
      end
    end

    context 'when the db ca cert is not configured' do
      let(:manifest_properties) do
        {
            "nfsv3driver" => {
                "tls" => {
                    "ca_cert" => nil,
                },
            }
        }
      end

      it 'successfully renders the script with an empty cert' do
        tpl_output = template.render(manifest_properties)

        expect(tpl_output).to eq("")
      end
    end
  end
end
