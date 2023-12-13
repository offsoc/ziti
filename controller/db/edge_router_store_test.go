package db

import (
	"github.com/google/uuid"
	"github.com/openziti/storage/boltz"
	"github.com/openziti/storage/boltztest"
	"testing"
)

func Test_EdgeRouterEvents(t *testing.T) {
	ctx := NewTestContext(t)
	defer ctx.Cleanup()
	ctx.Init()

	eventChecker := boltz.NewTestEventChecker(&ctx.Assertions)
	eventChecker.AddHandlers(ctx.stores.Router)
	eventChecker.AddHandlers(ctx.stores.EdgeRouter)

	fp := uuid.NewString()
	edgeRouter := &EdgeRouter{
		Router: Router{
			BaseExtEntity: boltz.BaseExtEntity{
				Id: uuid.NewString(),
			},
			Name:        uuid.NewString(),
			Fingerprint: &fp,
		},
	}

	boltztest.RequireCreate(ctx, edgeRouter)
	eventChecker.RequireEvent(boltz.TestEntityTypeParent, edgeRouter, boltz.EntityCreated)
	eventChecker.RequireEvent(boltz.TestEntityTypeChild, edgeRouter, boltz.EntityCreated)
	eventChecker.RequireNoEvent()

	// check events generated by updating from the edge-router store
	newFp := uuid.NewString()
	edgeRouter.Name = uuid.NewString()
	edgeRouter.Fingerprint = &newFp
	boltztest.RequireUpdate(ctx, edgeRouter)

	entity := eventChecker.RequireEvent(boltz.TestEntityTypeParent, edgeRouter, boltz.EntityUpdated)
	r, ok := entity.(*Router)
	ctx.True(ok)
	ctx.Equal(edgeRouter.Name, r.Name)
	ctx.NotNil(r.Fingerprint)
	ctx.Equal(*edgeRouter.Fingerprint, *r.Fingerprint)

	entity = eventChecker.RequireEvent(boltz.TestEntityTypeChild, edgeRouter, boltz.EntityUpdated)
	er, ok := entity.(*EdgeRouter)
	ctx.True(ok)
	ctx.Equal(edgeRouter.Name, er.Name)
	ctx.NotNil(er.Fingerprint)
	ctx.Equal(*edgeRouter.Fingerprint, *er.Fingerprint)

	eventChecker.RequireNoEvent()

	// check events generated by updating from the router store
	newFp = uuid.NewString()
	edgeRouter.Name = uuid.NewString()
	edgeRouter.Fingerprint = &newFp
	boltztest.RequireUpdate(ctx, &edgeRouter.Router)

	entity = eventChecker.RequireEvent(boltz.TestEntityTypeParent, edgeRouter, boltz.EntityUpdated)
	r, ok = entity.(*Router)
	ctx.True(ok)
	ctx.Equal(edgeRouter.Name, r.Name)
	ctx.NotNil(r.Fingerprint)
	ctx.Equal(*edgeRouter.Fingerprint, *r.Fingerprint)

	entity = eventChecker.RequireEvent(boltz.TestEntityTypeChild, edgeRouter, boltz.EntityUpdated)
	er, ok = entity.(*EdgeRouter)
	ctx.True(ok)
	ctx.Equal(edgeRouter.Name, er.Name)
	ctx.NotNil(er.Fingerprint)
	ctx.Equal(*edgeRouter.Fingerprint, *er.Fingerprint)

	// ensure setting fingerprint to nil works
	edgeRouter.Fingerprint = nil
	boltztest.RequireUpdate(ctx, edgeRouter)

	entity = eventChecker.RequireEvent(boltz.TestEntityTypeParent, edgeRouter, boltz.EntityUpdated)
	r, ok = entity.(*Router)
	ctx.True(ok)
	ctx.Equal(edgeRouter.Name, r.Name)
	ctx.Nil(edgeRouter.Fingerprint)
	ctx.Nil(r.Fingerprint)

	entity = eventChecker.RequireEvent(boltz.TestEntityTypeChild, edgeRouter, boltz.EntityUpdated)
	er, ok = entity.(*EdgeRouter)
	ctx.True(ok)
	ctx.Equal(edgeRouter.Name, er.Name)
	ctx.Nil(er.Fingerprint)

	// check delete events
	boltztest.RequireDelete(ctx, edgeRouter)
	eventChecker.RequireEvent(boltz.TestEntityTypeParent, edgeRouter, boltz.EntityDeleted)
	eventChecker.RequireEvent(boltz.TestEntityTypeChild, edgeRouter, boltz.EntityDeleted)
	eventChecker.RequireNoEvent()

	// check delete again, this time invoked from the child store
	fp = uuid.NewString()
	edgeRouter = &EdgeRouter{
		Router: Router{
			BaseExtEntity: boltz.BaseExtEntity{
				Id: uuid.NewString(),
			},
			Name:        uuid.NewString(),
			Fingerprint: &fp,
		},
	}

	boltztest.RequireCreate(ctx, edgeRouter)
	eventChecker.RequireEvent(boltz.TestEntityTypeParent, edgeRouter, boltz.EntityCreated)
	eventChecker.RequireEvent(boltz.TestEntityTypeChild, edgeRouter, boltz.EntityCreated)
	eventChecker.RequireNoEvent()

	boltztest.RequireDelete(ctx, edgeRouter)
	eventChecker.RequireEvent(boltz.TestEntityTypeParent, edgeRouter, boltz.EntityDeleted)
	eventChecker.RequireEvent(boltz.TestEntityTypeChild, edgeRouter, boltz.EntityDeleted)
	eventChecker.RequireNoEvent()
}