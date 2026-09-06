package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/alginugraha/monify/internal/domain"
)

type CategoryService struct {
	repo CategoryRepo
}

func NewCategoryService(repo CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

type CategoryInput struct {
	Name     string
	Kind     string
	ParentID string // "" for top-level
}

func (in CategoryInput) validate() error {
	if strings.TrimSpace(in.Name) == "" {
		return fmt.Errorf("%w: nama kategori wajib diisi", domain.ErrInvalid)
	}
	if !domain.CategoryKind(in.Kind).Valid() {
		return fmt.Errorf("%w: jenis kategori tidak valid", domain.ErrInvalid)
	}
	return nil
}

func parentPtr(id string) *string {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	return &id
}

func (s *CategoryService) List(ctx context.Context) ([]domain.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) Create(ctx context.Context, in CategoryInput) (domain.Category, error) {
	if err := in.validate(); err != nil {
		return domain.Category{}, err
	}
	return s.repo.Create(ctx, domain.Category{
		Name:     strings.TrimSpace(in.Name),
		Kind:     domain.CategoryKind(in.Kind),
		ParentID: parentPtr(in.ParentID),
	})
}

func (s *CategoryService) Update(ctx context.Context, id string, in CategoryInput) error {
	if err := in.validate(); err != nil {
		return err
	}
	cur, err := s.repo.ByID(ctx, id)
	if err != nil {
		return err
	}
	if in.ParentID == id {
		return fmt.Errorf("%w: kategori tidak boleh jadi induk dirinya sendiri", domain.ErrInvalid)
	}
	cur.Name = strings.TrimSpace(in.Name)
	cur.Kind = domain.CategoryKind(in.Kind)
	cur.ParentID = parentPtr(in.ParentID)
	return s.repo.Update(ctx, cur)
}

func (s *CategoryService) Delete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}
