package aggregate

import (
	"context"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/aggssert"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/algebra/aggregate/store"
)

type newAgg = func() aggssert.Aggregate

type HandleFunc = func(agg aggssert.Aggregate) error

type dataStore interface {
	ReadChanges(ctx context.Context, aggregateID string) ([]runtime.Change, error)
	AppendChanges(ctx context.Context, aggregateID string, version uint64, changes []runtime.Change) error
}

func NewAggregate(new newAgg, store dataStore) *storere {
	return &storere{
		new:   new,
		store: store,
	}
}

type storere struct {
	new   newAgg
	store dataStore
}

func (s *storere) NewAggregate(ctx context.Context, aggregateID string, handle HandleFunc) (aggssert.Aggregate, error) {
	fmt.Println("NewAggregate ID=" + aggregateID)
	_, err := s.store.ReadChanges(ctx, aggregateID)
	if err == nil {
		return nil, fmt.Errorf("NewAggregate, on aggregate that exits %s", aggregateID)
	} else if err != store.ErrNotFound {
		return nil, fmt.Errorf("NewAggregate, unknow error on aggregate %s. Detail: %w", aggregateID, err)
	}

	agg := s.new()
	err = handle(agg)
	if err != nil {
		return nil, fmt.Errorf("NewAggregate, error while mutating aggregate %s. Details: %w", aggregateID, err)
	}

	err = s.save(ctx, aggregateID, ^uint64(0), agg)
	if err != nil {
		return nil, err
	}

	return agg, nil
}

func (s *storere) MutateAggregate(ctx context.Context, aggregateID string, handle HandleFunc) (aggssert.Aggregate, error) {
	fmt.Println("MutateAggregate ID=" + aggregateID)
	var lastError error
	for retry := 0; retry < 2; retry++ {
		changes, err := s.store.ReadChanges(ctx, aggregateID)
		if err != nil {
			return nil, err
		}

		var version uint64 = 0
		agg := s.new()
		for _, change := range changes {
			version++
			err = agg.Changes().Append(change.Payload).Ok.ReduceRecent(agg).Err
			if err != nil {
				return nil, fmt.Errorf("MutateAggregate, error while replying aggregate %s, event=%#v. Details: %w", aggregateID, change, err)
			}
			version = change.Version
		}

		err = handle(agg)
		if err != nil {
			return nil, fmt.Errorf("MutateAggregate, error while mutating aggregate %s, event=%#v. Details: %w", aggregateID, err)
		}

		lastError = s.save(ctx, aggregateID, version, agg)
		if lastError == nil {
			return agg, nil
		}
	}

	return nil, fmt.Errorf("MutateAggregate, fail to store even after retrying X times. Details: %w", lastError)
}

func (s storere) save(ctx context.Context, aggregateID string, version uint64, agg aggssert.Aggregate) error {
	newChanges := make([]runtime.Change, 0)
	err := agg.Changes().ReduceChange(func(change runtime.Change, result *runtime.Reduced) *runtime.Reduced {
		if version == ^uint64(0) || change.Version > version {
			newChanges = append(newChanges, change)
		}
		return result
	}, nil).Err
	if err != nil {
		panic(fmt.Errorf("MutateAggregate, error while Reduce() changes must not happen. Details: %w", err))
	}

	return s.store.AppendChanges(ctx, aggregateID, version, newChanges)
}
