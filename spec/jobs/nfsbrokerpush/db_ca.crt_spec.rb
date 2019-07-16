require 'rspec'
require 'bosh/template/test'

describe 'nfsbrokerpush job' do
  let(:release) {Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..'))}
  let(:job) {release.job('nfsbrokerpush')}

  describe 'db_ca.crt.erb' do
    let(:template) {job.template('db_ca.crt')}

    context 'when configured with a db ca cert' do
      let(:manifest_properties) do
        {
            "nfsbrokerpush" => {
                "db" => {
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
            "nfsbrokerpush" => {
                "db" => {
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
