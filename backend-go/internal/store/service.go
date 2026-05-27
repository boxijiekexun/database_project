package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrCategoryInUse   = errors.New("category has dishes")
	ErrDuplicateName   = errors.New("duplicate name")
	ErrInvalidRelation = errors.New("invalid relation")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMerchantStore(ctx context.Context, userID int64) (Store, error) {
	item, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return Store{}, normalizeLookupError(err)
	}
	return item, nil
}

func (s *Service) UpdateMerchantStore(ctx context.Context, userID int64, input StoreUpdateInput) (Store, error) {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return Store{}, normalizeLookupError(err)
	}
	item, err := s.repo.UpdateStoreByID(ctx, storeItem.ID, input)
	if err != nil {
		return Store{}, normalizeDBError(err)
	}
	return item, nil
}

func (s *Service) ListMerchantCategories(ctx context.Context, userID int64) ([]DishCategory, error) {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return nil, normalizeLookupError(err)
	}
	return s.repo.ListCategoriesByStoreID(ctx, storeItem.ID)
}

func (s *Service) CreateMerchantCategory(ctx context.Context, userID int64, input CategoryCreateInput) (DishCategory, error) {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return DishCategory{}, normalizeLookupError(err)
	}
	item, err := s.repo.CreateCategory(ctx, storeItem.ID, input)
	if err != nil {
		return DishCategory{}, normalizeDBError(err)
	}
	return item, nil
}

func (s *Service) UpdateMerchantCategory(ctx context.Context, userID, categoryID int64, input CategoryUpdateInput) (DishCategory, error) {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return DishCategory{}, normalizeLookupError(err)
	}
	item, err := s.repo.UpdateCategory(ctx, storeItem.ID, categoryID, input)
	if err != nil {
		return DishCategory{}, normalizeDBError(err)
	}
	return item, nil
}

func (s *Service) DeleteMerchantCategory(ctx context.Context, userID, categoryID int64) error {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return normalizeLookupError(err)
	}
	count, err := s.repo.CountDishesByCategory(ctx, storeItem.ID, categoryID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCategoryInUse
	}
	if err := s.repo.DeleteCategory(ctx, storeItem.ID, categoryID); err != nil {
		return normalizeLookupError(err)
	}
	return nil
}

func (s *Service) ListMerchantDishes(ctx context.Context, userID int64, filter DishQueryFilter) ([]Dish, error) {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return nil, normalizeLookupError(err)
	}
	return s.repo.ListDishes(ctx, storeItem.ID, filter, false)
}

func (s *Service) CreateMerchantDish(ctx context.Context, userID int64, input DishCreateInput) (Dish, error) {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return Dish{}, normalizeLookupError(err)
	}
	if _, err := s.repo.GetCategoryByID(ctx, storeItem.ID, input.CategoryID); err != nil {
		return Dish{}, ErrInvalidRelation
	}
	item, err := s.repo.CreateDish(ctx, storeItem.ID, input)
	if err != nil {
		return Dish{}, normalizeDBError(err)
	}
	return item, nil
}

func (s *Service) UpdateMerchantDish(ctx context.Context, userID, dishID int64, input DishUpdateInput) (Dish, error) {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return Dish{}, normalizeLookupError(err)
	}
	if input.CategoryID != nil {
		if _, err := s.repo.GetCategoryByID(ctx, storeItem.ID, *input.CategoryID); err != nil {
			return Dish{}, ErrInvalidRelation
		}
	}
	item, err := s.repo.UpdateDish(ctx, storeItem.ID, dishID, input)
	if err != nil {
		return Dish{}, normalizeDBError(err)
	}
	return item, nil
}

func (s *Service) UpdateDishStatus(ctx context.Context, userID, dishID int64, status DishStatus) (Dish, error) {
	storeItem, err := s.repo.GetMerchantStoreByUserID(ctx, userID)
	if err != nil {
		return Dish{}, normalizeLookupError(err)
	}
	item, err := s.repo.UpdateDishStatus(ctx, storeItem.ID, dishID, status)
	if err != nil {
		return Dish{}, normalizeDBError(err)
	}
	return item, nil
}

func (s *Service) ListPublicStores(ctx context.Context) ([]Store, error) {
	return s.repo.ListPublicStores(ctx)
}

func (s *Service) GetPublicStore(ctx context.Context, storeID int64) (Store, error) {
	item, err := s.repo.GetStoreByID(ctx, storeID)
	if err != nil {
		return Store{}, normalizeLookupError(err)
	}
	return item, nil
}

func (s *Service) ListPublicCategories(ctx context.Context, storeID int64) ([]DishCategory, error) {
	if _, err := s.repo.GetStoreByID(ctx, storeID); err != nil {
		return nil, normalizeLookupError(err)
	}
	return s.repo.ListCategoriesByStoreID(ctx, storeID)
}

func (s *Service) ListPublicDishes(ctx context.Context, storeID int64, filter DishQueryFilter) ([]Dish, error) {
	if _, err := s.repo.GetStoreByID(ctx, storeID); err != nil {
		return nil, normalizeLookupError(err)
	}
	return s.repo.ListDishes(ctx, storeID, filter, true)
}

func (s *Service) ListPublicReviews(ctx context.Context, storeID int64) ([]Review, error) {
	if _, err := s.repo.GetStoreByID(ctx, storeID); err != nil {
		return nil, normalizeLookupError(err)
	}
	return s.repo.ListReviewsByStoreID(ctx, storeID)
}

func normalizeLookupError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func normalizeDBError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return ErrDuplicateName
		case 1452:
			return ErrInvalidRelation
		}
	}
	return err
}
