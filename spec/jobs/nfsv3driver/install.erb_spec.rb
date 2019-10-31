require 'rspec'
require 'bosh/template/test'

describe 'nfsv3driver job' do
  let(:release) {Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..'))}
  let(:job) {release.job('nfsv3driver')}

  describe 'ca.crt.erb' do
    let(:template) {job.template('bin/pre-start')}

    context 'when credhub has been set to zero instances' do
      let(:manifest_properties) do
        {
            "nfsv3driver" => {
                "enabled" => "true",
            }
        }
      end

      it 'succeeds' do
        expect{template.render(manifest_properties)}.not_to raise_error
      end
    end
  end
end
