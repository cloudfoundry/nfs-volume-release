require 'rspec'
require 'bosh/template/test'

describe 'nfsbroker-bbr-lock job' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('nfsbroker-bbr-lock') }

  let(:merged_manifest_properties) do
    {}
  end

  describe 'bbr metadata' do
    let(:template) { job.template('bin/bbr/metadata') }
    let(:links) { [] }

    subject(:script_output) do
      `#{template.render(merged_manifest_properties, consumes: links)}`
    end

    describe 'when the property is not defined' do
      it 'has a sane default' do
        expect(script_output).to eq('---
backup_should_be_locked_before:
- job_name: uaa
  release: uaa
- job_name: credhub
  release: credhub
- job_name: cloud_controller_ng
  release: capi
- job_name: cloud_controller_worker
  release: capi
- job_name: cc_deployment_updater
  release: capi

restore_should_be_locked_before:
- job_name: uaa
  release: uaa
- job_name: credhub
  release: credhub
- job_name: cloud_controller_ng
  release: capi
- job_name: cloud_controller_worker
  release: capi
- job_name: cc_deployment_updater
  release: capi

')
      end
    end
    describe 'when the property is provided' do
      before do
        merged_manifest_properties['bbr'] = {
          'metadata'=> '---
  test: yaml
  corn: banana
'
        }
      end
      it 'overrides the output' do
        expect(script_output).to eq('---
  test: yaml
  corn: banana

')
      end
    end
  end
end
