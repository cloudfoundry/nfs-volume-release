require 'rspec'
require 'bosh/template/test'

describe 'nfsv3driver job' do
  let(:release) {Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..'))}
  let(:job) {release.job('nfsv3driver')}

  describe 'ca.crt.erb' do
    let(:template) {job.template('bin/pre-start')}
    credhub_link = [
        Bosh::Template::Test::Link.new(
            name: 'credhub'
        )
    ]

    context 'when credhub has been set to zero instances' do
      let(:manifest_properties) do
        {
            "nfsv3driver" => {
                "enabled" => "true",
            }
        }
      end

      it 'a meaningful error message is returned' do
        expect{template.render(manifest_properties, consumes: credhub_link)}.to raise_error('credhub is required. Zero instances found.')
      end
    end
  end
end
