package main_test

import (
	. "code.cloudfoundry.org/nfsbroker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"
)

var _ = Describe("Services", func() {
	var (
		services Services
	)

	BeforeEach(func() {
		var err error
		services, err = NewServicesFromConfig("./test_default_services.json")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("List", func() {
		It("returns the list of services", func() {
			Expect(services.List()).To(Equal([]domain.Service{
				{
					ID:            "nfsbroker",
					Name:          "nfs-legacy",
					Description:   "Existing NFSv3 volumes (see: https://code.cloudfoundry.org/nfs-volume-release/)",
					Bindable:      true,
					PlanUpdatable: false,
					Tags:          []string{"nfs"},
					Requires:      []domain.RequiredPermission{"volume_mount"},

					Plans: []domain.ServicePlan{
						{
							Name:        "Existing",
							ID:          "Existing",
							Description: "A preexisting filesystem",
						},
					},
				},
				{
					ID:            "997f8f26-e10c-11e7-80c1-9a214cf093ae",
					Name:          "nfs",
					Description:   "Existing NFSv3 and v4 volumes",
					Bindable:      true,
					PlanUpdatable: false,
					Tags:          []string{"nfs"},
					Requires:      []domain.RequiredPermission{"volume_mount"},

					Plans: []domain.ServicePlan{
						{
							Name:        "Existing",
							ID:          "09a09260-1df5-4445-9ed7-1ba56dadbbc8",
							Description: "A preexisting filesystem",
						},
					},
				},
			}))
		})
	})
})
