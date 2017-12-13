package bosh_test

import (
	"github.com/pivotal-cf-experimental/om-manifest-validator/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {

	var (
		manifest *bosh.Manifest
	)

	Describe("InstanceGroups FindJob", func() {
		Context("when the InstanceGroup has a Jobs section", func() {
			var instanceGroup *bosh.InstanceGroup
			BeforeEach(func() {
				job := bosh.NewJob("existentJob-partition-random-guid")
				instanceGroup = bosh.NewInstanceGroup("existentInstanceGroup", []*bosh.Job{job})
			})

			It("returns a Job matching the given name", func() {
				expectedJob := instanceGroup.FindJob("existentJob-partition-random-guid")
				Expect(expectedJob.Name()).To(HavePrefix("existentJob"))
			})

			It("panics when no match is found", func() {
				Expect(func() { manifest.JobNamed("nonExistentJob") }).To(Panic())
			})
		})
	})

	Describe("InstanceGroupNamed", func() {
		Context("when the manifest has a InstanceGroups section", func() {
			BeforeEach(func() {
				instanceGroup := bosh.NewInstanceGroup("existentIG-random-guid")
				manifest = &bosh.Manifest{
					InstanceGroups: []*bosh.InstanceGroup{instanceGroup},
				}
			})

			It("returns an InstanceGroup matching the given name", func() {
				expectedJob := manifest.InstanceGroupNamed("existentIG-random-guid")
				Expect(expectedJob.Name()).To(HavePrefix("existentIG-random-guid"))
			})

			It("return nil when no match is found", func() {
				Expect(manifest.InstanceGroupNamed("nonExistentJob")).To(BeNil())
			})
		})
	})


	Describe("MustFindInstanceGroupNamed", func() {
		Context("when the manifest has a InstanceGroups section", func() {
			BeforeEach(func() {
				instanceGroup := bosh.NewInstanceGroup("existentIG-random-guid")
				manifest = &bosh.Manifest{
					InstanceGroups: []*bosh.InstanceGroup{instanceGroup},
				}
			})

			It("returns an InstanceGroup matching the given name", func() {
				expectedJob := manifest.MustFindInstanceGroupNamed("existentIG-random-guid")
				Expect(expectedJob.Name()).To(HavePrefix("existentIG-random-guid"))
			})

			It("panics when no match is found", func() {
				Expect(func() { manifest.MustFindInstanceGroupNamed("nonExistentJob") }).To(Panic())
			})
		})
	})

	Describe("JobNamed", func() {
		Context("when the manifest has a Jobs section", func() {
			BeforeEach(func() {
				job := bosh.NewJob("existentJob-partition-random-guid")
				manifest = &bosh.Manifest{
					Jobs: []*bosh.Job{job},
				}
			})

			It("returns a Job matching the given name", func() {
				expectedJob := manifest.JobNamed("existentJob")
				Expect(expectedJob.Name()).To(HavePrefix("existentJob"))

			})

			It("panics when no match is found", func() {
				Expect(func() { manifest.JobNamed("nonExistentJob") }).To(Panic())
			})
		})

		Context("when the manifest does not have a Jobs section", func() {
			BeforeEach(func() {
				instanceGroup := bosh.NewInstanceGroup("existentInstanceGroup")
				manifest = &bosh.Manifest{
					InstanceGroups: []*bosh.InstanceGroup{instanceGroup},
				}
			})

			It("returns an InstanceGroup matching the given name", func() {
				expectedInstanceGroup := manifest.JobNamed("existentInstanceGroup")
				Expect(expectedInstanceGroup.Name()).To(Equal("existentInstanceGroup"))
			})

			It("panics when no match is found", func() {
				Expect(func() { manifest.JobNamed("nonExistentInstanceGroup") }).To(Panic())
			})
		})
	})

	Describe("Find", func() {
		Context("when the lens is un-nested", func() {
			Context("and the property is not present", func() {
				It("returns an error", func() {
					p := &bosh.Properties{}
					v, err := p.Find("nonExistentProperty")
					Expect(v).To(BeNil())
					Expect(err).To(HaveOccurred())
				})
			})
			Context("and the property is present", func() {
				It("returns the property value", func() {
					p := &bosh.Properties{
						"anotherProperty":  "foo",
						"existentProperty": "some-value",
					}
					v, err := p.Find("existentProperty")
					Expect(v).To(Equal("some-value"))
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
		Context("when the lens is nested", func() {
			Context("and the property is not present", func() {
				It("returns an error", func() {
					p := &bosh.Properties{
						"anotherProperty": "foo",
					}
					v, err := p.Find("a.nonExistentProperty")
					Expect(v).To(BeNil())
					Expect(err).To(HaveOccurred())
				})
			})
			Context("and the property is present", func() {
				It("returns the property value", func() {
					p := &bosh.Properties{
						"anotherProperty": "foo",
						"an": bosh.Properties{
							"existentProperty": "some-value",
						},
					}
					v, err := p.Find("an.existentProperty")
					Expect(v).To(Equal("some-value"))
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
		Context("when the lens is deeply nested", func() {
			Context("and the property is not present", func() {
				It("returns an error", func() {
					p := &bosh.Properties{
						"anotherProperty": "foo",
						"an": bosh.Properties{
							"alternative": bosh.Properties{
								"property": "another-value",
							},
						},
					}
					v, err := p.Find("an.alternative.nonExistent.property")
					Expect(v).To(BeNil())
					Expect(err).To(HaveOccurred())
				})
			})
			Context("and the property is present", func() {
				It("returns the property value", func() {
					p := &bosh.Properties{
						"anotherProperty": "foo",
						"a": bosh.Properties{
							"deeply": bosh.Properties{
								"nested": bosh.Properties{
									"existentProperty": "some-value",
								},
							},
						},
					}
					v, err := p.Find("a.deeply.nested.existentProperty")
					Expect(v).To(Equal("some-value"))
					Expect(err).ToNot(HaveOccurred())
				})
			})
			Context("and an intermediate node is not of the expected type", func() {
				It("panics", func() {
					p := &bosh.Properties{
						"anotherProperty": "foo",
						"an": map[string]string{
							"unusual": "property",
						},
					}
					Expect(func() { p.Find("an.unusual.property") }).To(Panic())
				})
			})
		})
		Context("when the value is an array", func() {
			It("returns the property value", func() {
				p := &bosh.Properties{
					"a": bosh.Properties{
						"nested": bosh.Properties{
							"array": []string{"foo", "bar", "baz"},
						},
					},
				}
				v, err := p.Find("a.nested.array")
				Expect(v).To(Equal([]string{"foo", "bar", "baz"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("FindString", func() {
		Context("when the property is not present", func() {
			It("returns an error", func() {
				p := &bosh.Properties{
					"a": bosh.Properties{},
				}
				v, err := p.FindString("a.nonExistent.property")
				Expect(v).To(BeZero())
				Expect(err).To(HaveOccurred())
			})
		})
		Context("when the property is an integer", func() {
			It("returns an error", func() {
				p := &bosh.Properties{
					"an": bosh.Properties{
						"integerProperty": 31,
					},
				}
				v, err := p.FindString("an.integerProperty")
				Expect(v).To(BeZero())
				Expect(err).To(HaveOccurred())
			})
		})
		Context("when the property is a string", func() {
			It("returns the value", func() {
				p := &bosh.Properties{
					"a": bosh.Properties{
						"stringProperty": "thirty-one",
					},
				}
				v, err := p.FindString("a.stringProperty")
				Expect(v).To(Equal("thirty-one"))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("FindInt", func() {
		Context("when the property is not present", func() {
			It("returns an error", func() {
				p := &bosh.Properties{
					"a": bosh.Properties{},
				}
				v, err := p.FindInt("a.nonExistent.property")
				Expect(v).To(BeZero())
				Expect(err).To(HaveOccurred())
			})
		})
		Context("when the property is an integer", func() {
			It("returns the value", func() {
				p := &bosh.Properties{
					"an": bosh.Properties{
						"integerProperty": 31,
					},
				}
				v, err := p.FindInt("an.integerProperty")
				Expect(v).To(Equal(31))
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("when the property is a string", func() {
			It("returns an error", func() {
				p := &bosh.Properties{
					"a": bosh.Properties{
						"stringProperty": "thirty-one",
					},
				}
				v, err := p.FindInt("a.stringProperty")
				Expect(v).To(BeZero())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("FindBool", func() {
		Context("when the property is not present", func() {
			It("returns an error", func() {
				p := &bosh.Properties{
					"a": bosh.Properties{},
				}
				v, err := p.FindBool("a.nonExistent.property")
				Expect(v).To(BeZero())
				Expect(err).To(HaveOccurred())
			})
		})
		Context("when the property is a boolean", func() {
			It("returns the value", func() {
				p := &bosh.Properties{
					"a": bosh.Properties{
						"booleanProperty": false,
					},
				}
				v, err := p.FindBool("a.booleanProperty")
				Expect(v).To(BeFalse())
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("when the property is a string", func() {
			It("returns an error", func() {
				p := &bosh.Properties{
					"a": bosh.Properties{
						"stringProperty": "true",
					},
				}
				v, err := p.FindBool("a.stringProperty")
				Expect(v).To(BeZero())
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
