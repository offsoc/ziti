package persistence

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/netfoundry/ziti-foundation/util/stringz"
	"go.etcd.io/bbolt"
	"sort"
	"testing"
)

func Test_ServicePolicyStore(t *testing.T) {
	ctx := NewTestContext(t)
	defer ctx.Cleanup()
	ctx.Init()

	t.Run("test create service policies", ctx.testCreateServicePolicy)
	t.Run("test create/update service policies with invalid entity refs", ctx.testServicePolicyInvalidValues)
	t.Run("test service policy evaluation", ctx.testServicePolicyRoleEvaluation)
	t.Run("test update/delete referenced entities", ctx.testServicePolicyUpdateDeleteRefs)
}

func (ctx *TestContext) testCreateServicePolicy(_ *testing.T) {
	ctx.cleanupAll()

	policy := newServicePolicy(uuid.New().String())
	ctx.requireCreate(policy)

	err := ctx.GetDb().View(func(tx *bbolt.Tx) error {
		ctx.validateBaseline(policy)
		ctx.Equal(0, len(ctx.stores.ServicePolicy.GetRelatedEntitiesIdList(tx, policy.Id, EntityTypeServices)))
		ctx.Equal(0, len(ctx.stores.ServicePolicy.GetRelatedEntitiesIdList(tx, policy.Id, EntityTypeIdentities)))

		testPolicy, err := ctx.stores.ServicePolicy.LoadOneByName(tx, policy.Name)
		ctx.NoError(err)
		ctx.NotNil(testPolicy)
		ctx.Equal(policy.Name, testPolicy.Name)

		return nil
	})
	ctx.NoError(err)
}

func (ctx *TestContext) testServicePolicyInvalidValues(_ *testing.T) {
	ctx.cleanupAll()

	// test identity roles
	policy := newServicePolicy(uuid.New().String())
	invalidId := uuid.New().String()
	policy.IdentityRoles = []string{entityRef(invalidId)}
	err := ctx.create(policy)
	ctx.EqualError(err, fmt.Sprintf("the value '[%v]' for 'identityRoles' is invalid: no identities found with the given names/ids", invalidId))

	policy.IdentityRoles = []string{AllRole, roleRef("other")}
	err = ctx.create(policy)
	ctx.EqualError(err, fmt.Sprintf("the value '[%v %v]' for 'identityRoles' is invalid: if using %v, it should be the only role specified", AllRole, roleRef("other"), AllRole))

	identityTypeId := ctx.getIdentityTypeId()
	identity := newIdentity(uuid.New().String(), identityTypeId)
	ctx.requireCreate(identity)

	policy.IdentityRoles = []string{entityRef(identity.Id), entityRef(invalidId)}
	err = ctx.create(policy)
	ctx.EqualError(err, fmt.Sprintf("the value '[%v]' for 'identityRoles' is invalid: no identities found with the given names/ids", invalidId))

	policy.IdentityRoles = []string{entityRef(identity.Id)}
	ctx.requireCreate(policy)
	ctx.validateServicePolicyIdentities([]*Identity{identity}, []*ServicePolicy{policy})
	ctx.requireDelete(policy)

	policy.IdentityRoles = []string{entityRef(identity.Name)}
	ctx.requireCreate(policy)
	ctx.validateServicePolicyIdentities([]*Identity{identity}, []*ServicePolicy{policy})

	policy.IdentityRoles = append(policy.IdentityRoles, entityRef(invalidId))
	err = ctx.update(policy)
	ctx.EqualError(err, fmt.Sprintf("the value '[%v]' for 'identityRoles' is invalid: no identities found with the given names/ids", invalidId))
	ctx.requireDelete(policy)

	// test service roles
	policy.IdentityRoles = nil
	policy.ServiceRoles = []string{entityRef(invalidId)}
	err = ctx.create(policy)
	ctx.EqualError(err, fmt.Sprintf("the value '[%v]' for 'serviceRoles' is invalid: no services found with the given names/ids", invalidId))

	policy.ServiceRoles = []string{AllRole, roleRef("other")}
	err = ctx.create(policy)
	ctx.EqualError(err, fmt.Sprintf("the value '[%v %v]' for 'serviceRoles' is invalid: if using %v, it should be the only role specified", AllRole, roleRef("other"), AllRole))

	service := newEdgeService(uuid.New().String())
	ctx.requireCreate(service)

	policy.ServiceRoles = []string{entityRef(service.Id), entityRef(invalidId)}
	err = ctx.create(policy)
	ctx.EqualError(err, fmt.Sprintf("the value '[%v]' for 'serviceRoles' is invalid: no services found with the given names/ids", invalidId))

	policy.ServiceRoles = []string{entityRef(service.Id)}
	ctx.requireCreate(policy)
	ctx.validateServicePolicyServices([]*EdgeService{service}, []*ServicePolicy{policy})
	ctx.requireDelete(policy)

	policy.ServiceRoles = []string{entityRef(service.Name)}
	ctx.requireCreate(policy)
	ctx.validateServicePolicyServices([]*EdgeService{service}, []*ServicePolicy{policy})

	policy.ServiceRoles = append(policy.ServiceRoles, entityRef(invalidId))
	err = ctx.update(policy)
	ctx.EqualError(err, fmt.Sprintf("the value '[%v]' for 'serviceRoles' is invalid: no services found with the given names/ids", invalidId))
	ctx.requireDelete(policy)
}

func (ctx *TestContext) testServicePolicyUpdateDeleteRefs(_ *testing.T) {
	ctx.cleanupAll()

	// test identity roles
	policy := newServicePolicy(uuid.New().String())
	identityTypeId := ctx.getIdentityTypeId()
	identity := newIdentity(uuid.New().String(), identityTypeId)
	ctx.requireCreate(identity)

	policy.IdentityRoles = []string{entityRef(identity.Id)}
	ctx.requireCreate(policy)
	ctx.validateServicePolicyIdentities([]*Identity{identity}, []*ServicePolicy{policy})
	ctx.requireDelete(identity)
	ctx.requireReload(policy)
	ctx.Equal(0, len(policy.IdentityRoles), "identity id should have been removed from identity roles")

	identity = newIdentity(uuid.New().String(), identityTypeId)
	ctx.requireCreate(identity)

	policy.IdentityRoles = []string{entityRef(identity.Name)}
	ctx.requireUpdate(policy)
	ctx.validateServicePolicyIdentities([]*Identity{identity}, []*ServicePolicy{policy})

	identity.Name = uuid.New().String()
	ctx.requireUpdate(identity)
	ctx.requireReload(policy)
	ctx.True(stringz.Contains(policy.IdentityRoles, entityRef(identity.Name)))
	ctx.validateServicePolicyIdentities([]*Identity{identity}, []*ServicePolicy{policy})

	ctx.requireDelete(identity)
	ctx.requireReload(policy)
	ctx.Equal(0, len(policy.IdentityRoles), "identity name should have been removed from identity roles")

	// test service roles
	service := newEdgeService(uuid.New().String())
	ctx.requireCreate(service)

	policy.ServiceRoles = []string{entityRef(service.Id)}
	ctx.requireUpdate(policy)
	ctx.validateServicePolicyServices([]*EdgeService{service}, []*ServicePolicy{policy})
	ctx.requireDelete(service)
	ctx.requireReload(policy)
	ctx.Equal(0, len(policy.ServiceRoles), "service id should have been removed from service roles")

	service = newEdgeService(uuid.New().String())
	ctx.requireCreate(service)

	policy.ServiceRoles = []string{entityRef(service.Name)}
	ctx.requireUpdate(policy)
	ctx.validateServicePolicyServices([]*EdgeService{service}, []*ServicePolicy{policy})

	service.Name = uuid.New().String()
	ctx.requireUpdate(service)
	ctx.requireReload(policy)
	ctx.True(stringz.Contains(policy.ServiceRoles, entityRef(service.Name)))
	ctx.validateServicePolicyServices([]*EdgeService{service}, []*ServicePolicy{policy})

	ctx.requireDelete(service)
	ctx.requireReload(policy)
	ctx.Equal(0, len(policy.ServiceRoles), "service name should have been removed from service roles")
}

func (ctx *TestContext) testServicePolicyRoleEvaluation(_ *testing.T) {
	ctx.cleanupAll()

	// create some identities, edge routers for reference by id
	// create initial policies, check state
	// create edge routers/identities with roles on create, check state
	// delete all er/identities, check state
	// create edge routers/identities with roles added after create, check state
	// add 5 new policies, check
	// modify polices, add roles, check
	// modify policies, remove roles, check

	identityTypeId := ctx.getIdentityTypeId()

	var identities []*Identity
	for i := 0; i < 5; i++ {
		identity := newIdentity(uuid.New().String(), identityTypeId)
		ctx.requireCreate(identity)
		identities = append(identities, identity)
	}

	var services []*EdgeService
	for i := 0; i < 5; i++ {
		service := newEdgeService(uuid.New().String())
		ctx.requireCreate(service)
		services = append(services, service)
	}

	identityRolesAttrs := []string{"foo", "bar", uuid.New().String(), "baz", uuid.New().String(), "quux"}
	var identityRoles []string
	for _, role := range identityRolesAttrs {
		identityRoles = append(identityRoles, roleRef(role))
	}

	serviceRoleAttrs := []string{uuid.New().String(), "another-role", "parsley, sage, rosemary and don't forget thyme", uuid.New().String(), "blop", "asdf"}
	var serviceRoles []string
	for _, role := range serviceRoleAttrs {
		serviceRoles = append(serviceRoles, roleRef(role))
	}

	multipleIdentityList := []string{identities[1].Id, identities[2].Id, identities[3].Id}
	multipleServiceList := []string{services[1].Id, services[2].Id, services[3].Id}

	policies := ctx.createServicePolicies(identityRoles, serviceRoles, identities, services, true)

	for i := 0; i < 9; i++ {
		relatedServices := ctx.getRelatedIds(policies[i], EntityTypeServices)
		relatedIdentities := ctx.getRelatedIds(policies[i], EntityTypeIdentities)
		if i == 3 {
			ctx.Equal([]string{services[0].Id}, relatedServices)
			ctx.Equal([]string{identities[0].Id}, relatedIdentities)
		} else if i == 4 || i == 5 {
			sort.Strings(multipleServiceList)
			sort.Strings(multipleIdentityList)
			ctx.Equal(multipleServiceList, relatedServices)
			ctx.Equal(multipleIdentityList, relatedIdentities)
		} else if i == 6 {
			ctx.Equal(5, len(relatedServices))
			ctx.Equal(5, len(relatedIdentities))
		} else {
			ctx.Equal(0, len(relatedIdentities))
			ctx.Equal(0, len(relatedServices))
		}
	}

	// no roles
	identity := newIdentity(uuid.New().String(), identityTypeId)
	ctx.requireCreate(identity)
	identities = append(identities, identity)

	stringz.Permutations(identityRolesAttrs, func(roles []string) {
		identity := newIdentity(uuid.New().String(), identityTypeId, roles...)
		ctx.requireCreate(identity)
		identities = append(identities, identity)
	})

	// no roles
	service := newEdgeService(uuid.New().String())
	ctx.requireCreate(service)
	services = append(services, service)

	stringz.Permutations(serviceRoleAttrs, func(roles []string) {
		service := newEdgeService(uuid.New().String(), roles...)
		ctx.requireCreate(service)
		services = append(services, service)
	})

	ctx.validateServicePolicyIdentities(identities, policies)
	ctx.validateServicePolicyServices(services, policies)

	for _, identity := range identities {
		ctx.requireDelete(identity)
	}

	for _, service := range services {
		ctx.requireDelete(service)
	}

	identities = nil
	services = nil

	stringz.Permutations(identityRolesAttrs, func(roles []string) {
		identity := newIdentity(uuid.New().String(), identityTypeId)
		ctx.requireCreate(identity)
		identity.RoleAttributes = roles
		ctx.requireUpdate(identity)
		identities = append(identities, identity)
	})

	stringz.Permutations(serviceRoleAttrs, func(roles []string) {
		service := newEdgeService(uuid.New().String())
		ctx.requireCreate(service)
		service.RoleAttributes = roles
		ctx.requireUpdate(service)
		services = append(services, service)
	})

	ctx.validateServicePolicyIdentities(identities, policies)
	ctx.validateServicePolicyServices(services, policies)

	// ensure policies get cleaned up
	for _, policy := range policies {
		ctx.requireDelete(policy)
	}

	// test with policies created after identities/edge routers
	policies = ctx.createServicePolicies(identityRoles, serviceRoles, identities, services, true)

	ctx.validateServicePolicyIdentities(identities, policies)
	ctx.validateServicePolicyServices(services, policies)

	for _, policy := range policies {
		ctx.requireDelete(policy)
	}

	// test with policies created after identities/edge routers and roles added after created
	policies = ctx.createServicePolicies(identityRoles, serviceRoles, identities, services, false)

	ctx.validateServicePolicyIdentities(identities, policies)
	ctx.validateServicePolicyServices(services, policies)

	for _, identity := range identities {
		if len(identity.RoleAttributes) > 0 {
			identity.RoleAttributes = identity.RoleAttributes[1:]
			ctx.requireUpdate(identity)
		}
	}

	for _, service := range services {
		if len(service.RoleAttributes) > 0 {
			service.RoleAttributes = service.RoleAttributes[1:]
			ctx.requireUpdate(service)
		}
	}

	for _, policy := range policies {
		if len(policy.IdentityRoles) > 0 {
			policy.IdentityRoles = policy.IdentityRoles[1:]
		}
		if len(policy.ServiceRoles) > 0 {
			policy.ServiceRoles = policy.ServiceRoles[1:]
		}
		ctx.requireUpdate(policy)
	}

	ctx.validateServicePolicyIdentities(identities, policies)
	ctx.validateServicePolicyServices(services, policies)
}

func (ctx *TestContext) createServicePolicies(identityRoles, serviceRoles []string, identities []*Identity, services []*EdgeService, oncreate bool) []*ServicePolicy {
	var policies []*ServicePolicy
	for i := 0; i < 9; i++ {
		policy := newServicePolicy(uuid.New().String())
		if !oncreate {
			ctx.requireCreate(policy)
		}
		if i == 1 {
			policy.IdentityRoles = []string{identityRoles[0]}
			policy.ServiceRoles = []string{serviceRoles[0]}
		}
		if i == 2 {
			policy.IdentityRoles = []string{identityRoles[1], identityRoles[2], identityRoles[3]}
			policy.ServiceRoles = []string{serviceRoles[1], serviceRoles[2], serviceRoles[3]}
		}
		if i == 3 {
			policy.IdentityRoles = []string{entityRef(identities[0].Id)}
			policy.ServiceRoles = []string{entityRef(services[0].Id)}
		}
		if i == 4 {
			policy.IdentityRoles = []string{entityRef(identities[1].Id), entityRef(identities[2].Name), entityRef(identities[3].Id)}
			policy.ServiceRoles = []string{entityRef(services[1].Id), entityRef(services[2].Name), entityRef(services[3].Id)}
		}
		if i == 5 {
			policy.IdentityRoles = []string{identityRoles[4], entityRef(identities[1].Id), entityRef(identities[2].Id), entityRef(identities[3].Name)}
			policy.ServiceRoles = []string{serviceRoles[4], entityRef(services[1].Id), entityRef(services[2].Id), entityRef(services[3].Name)}
		}
		if i == 6 {
			policy.IdentityRoles = []string{AllRole}
			policy.ServiceRoles = []string{AllRole}
		}
		if i == 7 {
			policy.Semantic = SemanticAnyOf
			policy.IdentityRoles = []string{identityRoles[0]}
			policy.ServiceRoles = []string{serviceRoles[0]}
		}
		if i == 8 {
			policy.Semantic = SemanticAnyOf
			policy.IdentityRoles = []string{identityRoles[1], identityRoles[2], identityRoles[3]}
			policy.ServiceRoles = []string{serviceRoles[1], serviceRoles[2], serviceRoles[3]}
		}

		policies = append(policies, policy)
		if oncreate {
			ctx.requireCreate(policy)
		} else {
			ctx.requireUpdate(policy)
		}
	}
	return policies
}

func (ctx *TestContext) validateServicePolicyIdentities(identities []*Identity, policies []*ServicePolicy) {
	for _, policy := range policies {
		count := 0
		relatedIdentities := ctx.getRelatedIds(policy, EntityTypeIdentities)
		for _, identity := range identities {
			relatedPolicies := ctx.getRelatedIds(identity, EntityTypeServicePolicies)
			shouldContain := ctx.policyShouldMatch(policy.Semantic, policy.IdentityRoles, identity, identity.RoleAttributes)

			policyContains := stringz.Contains(relatedIdentities, identity.Id)
			ctx.Equal(shouldContain, policyContains, "entity roles attr: %v. policy roles: %v", identity.RoleAttributes, policy.IdentityRoles)
			if shouldContain {
				count++
			}

			entityContains := stringz.Contains(relatedPolicies, policy.Id)
			ctx.Equal(shouldContain, entityContains, "identity: %v, policy: %v, entity roles attr: %v. policy roles: %v",
				identity.Id, policy.Id, identity.RoleAttributes, policy.IdentityRoles)
		}
		ctx.Equal(count, len(relatedIdentities))
	}
}

func (ctx *TestContext) validateServicePolicyServices(services []*EdgeService, policies []*ServicePolicy) {
	for _, policy := range policies {
		count := 0
		relatedServices := ctx.getRelatedIds(policy, EntityTypeServices)
		for _, service := range services {
			relatedPolicies := ctx.getRelatedIds(service, EntityTypeServicePolicies)
			shouldContain := ctx.policyShouldMatch(policy.Semantic, policy.ServiceRoles, service, service.RoleAttributes)
			policyContains := stringz.Contains(relatedServices, service.Id)
			ctx.Equal(shouldContain, policyContains, "entity roles attr: %v. policy roles: %v", service.RoleAttributes, policy.ServiceRoles)
			if shouldContain {
				count++
			}

			entityContains := stringz.Contains(relatedPolicies, policy.Id)
			ctx.Equal(shouldContain, entityContains, "identity: %v, policy: %v, entity roles attr: %v. policy roles: %v",
				service.Id, policy.Id, service.RoleAttributes, policy.ServiceRoles)
		}
		ctx.Equal(count, len(relatedServices))
	}
}
