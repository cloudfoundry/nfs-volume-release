require 'rspec'
require 'json'
require 'bosh/template/test'

describe 'nfsbrokerpush job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('nfsbrokerpush') }

  describe 'config/services.json' do
    let(:template) { job.template('config/services.json') }

    context 'with the default services' do
      it 'renders the catalog metadata without the optional marketplace fields' do
        services = JSON.parse(template.render({}))

        expect(services.length).to eq(1)
        metadata = services[0]['metadata']
        expect(metadata.keys).to contain_exactly(
          'displayName', 'imageUrl', 'longDescription', 'providerDisplayName', 'documentationUrl', 'supportType'
        )
        expect(metadata['displayName']).to eq('NFS V3 / V4 Volume Broker')
      end
    end

    context 'when a service sets types, categories and verified' do
      let(:manifest_properties) do
        {
          'nfsbrokerpush' => {
            'services' => [
              {
                'id' => 'some-service-id',
                'name' => 'nfs',
                'description' => 'some-description',
                'bindable' => true,
                'plan_updatable' => false,
                'tags' => ['nfs'],
                'requires' => ['volume_mount'],
                'metadata' => {
                  'display_name' => 'some-display-name',
                  'types' => ['SERVICE'],
                  'categories' => ['PLATFORM_AND_DELIVERY'],
                  'verified' => true,
                },
                'plans' => [
                  {
                    'id' => 'some-plan-id',
                    'name' => 'Existing',
                    'description' => 'some-plan-description',
                  }
                ],
              }
            ]
          }
        }
      end

      it 'passes them through to the catalog metadata' do
        metadata = JSON.parse(template.render(manifest_properties))[0]['metadata']

        expect(metadata['displayName']).to eq('some-display-name')
        expect(metadata['types']).to eq(['SERVICE'])
        expect(metadata['categories']).to eq(['PLATFORM_AND_DELIVERY'])
        expect(metadata['verified']).to eq(true)
      end

      it 'passes verified through when it is false' do
        manifest_properties['nfsbrokerpush']['services'][0]['metadata']['verified'] = false

        metadata = JSON.parse(template.render(manifest_properties))[0]['metadata']

        expect(metadata).to include('verified' => false)
      end
    end
  end
end
