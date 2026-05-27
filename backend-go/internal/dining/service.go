package dining

import (
	"context"
	"errors"

	"project1/internal/store"
)

var ErrDiningNotFound = errors.New("dining not found")

type Service struct {
	repo      *Repository
	storeRepo *store.Repository
}

func NewService(repo *Repository, storeRepo *store.Repository) *Service {
	return &Service{repo: repo, storeRepo: storeRepo}
}

func (s *Service) ListAreas(ctx context.Context) ([]DiningArea, error) {
	return s.repo.ListAreas(ctx)
}

func (s *Service) ListCanteens(ctx context.Context, areaID int64) ([]Canteen, error) {
	if _, err := s.repo.GetAreaByID(ctx, areaID); err != nil {
		return nil, err
	}
	return s.repo.ListCanteensByAreaID(ctx, areaID)
}

func (s *Service) GetCanteen(ctx context.Context, canteenID int64) (Canteen, error) {
	item, err := s.repo.GetCanteenByID(ctx, canteenID)
	if err != nil {
		return Canteen{}, err
	}
	return item, nil
}

func (s *Service) ListStores(ctx context.Context, canteenID int64) ([]store.Store, error) {
	if _, err := s.repo.GetCanteenByID(ctx, canteenID); err != nil {
		return nil, err
	}
	return s.storeRepo.ListStoresByCanteenID(ctx, canteenID)
}

func (s *Service) ListDishes(ctx context.Context, canteenID int64, filter store.DishQueryFilter) ([]store.Dish, error) {
	if _, err := s.repo.GetCanteenByID(ctx, canteenID); err != nil {
		return nil, err
	}
	return s.storeRepo.ListDishesByCanteenID(ctx, canteenID, filter, true)
}
