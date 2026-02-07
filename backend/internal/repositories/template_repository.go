package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	db "wish-list/internal/db/models"
)

// Sentinel errors for template repository
var (
	ErrTemplateNotFound        = errors.New("template not found")
	ErrDefaultTemplateNotFound = errors.New("default template not found")
)

// TemplateRepositoryInterface defines the interface for template database operations
type TemplateRepositoryInterface interface {
	GetByID(ctx context.Context, id string) (*db.Template, error)
	GetAll(ctx context.Context) ([]*db.Template, error)
	GetDefault(ctx context.Context) (*db.Template, error)
	Create(ctx context.Context, template db.Template) (*db.Template, error)
	Update(ctx context.Context, template db.Template) (*db.Template, error)
	Delete(ctx context.Context, id string) error
}

type TemplateRepository struct {
	db *db.DB
}

func NewTemplateRepository(database *db.DB) TemplateRepositoryInterface {
	return &TemplateRepository{
		db: database,
	}
}

// GetByID retrieves a template by ID
func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*db.Template, error) {
	query := `
		SELECT
			id, name, description, preview_image_url, config, is_default, created_at, updated_at
		FROM templates
		WHERE id = $1
	`

	var template db.Template
	err := r.db.GetContext(ctx, &template, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &template, nil
}

// GetAll retrieves all templates
func (r *TemplateRepository) GetAll(ctx context.Context) ([]*db.Template, error) {
	query := `
		SELECT
			id, name, description, preview_image_url, config, is_default, created_at, updated_at
		FROM templates
		ORDER BY created_at DESC
	`

	var templates []*db.Template
	err := r.db.SelectContext(ctx, &templates, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all templates: %w", err)
	}

	return templates, nil
}

// GetDefault retrieves the default template
func (r *TemplateRepository) GetDefault(ctx context.Context) (*db.Template, error) {
	query := `
		SELECT
			id, name, description, preview_image_url, config, is_default, created_at, updated_at
		FROM templates
		WHERE is_default = true
		LIMIT 1
	`

	var template db.Template
	err := r.db.GetContext(ctx, &template, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDefaultTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get default template: %w", err)
	}

	return &template, nil
}

// Create inserts a new template into the database
func (r *TemplateRepository) Create(ctx context.Context, template db.Template) (*db.Template, error) {
	query := `
		INSERT INTO templates (
			id, name, description, preview_image_url, config, is_default
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING
			id, name, description, preview_image_url, config, is_default, created_at, updated_at
	`

	var createdTemplate db.Template
	err := r.db.QueryRowxContext(ctx, query,
		template.ID,
		template.Name,
		db.TextToString(template.Description),
		db.TextToString(template.PreviewImageUrl),
		template.Config,
		template.IsDefault,
	).StructScan(&createdTemplate)

	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return &createdTemplate, nil
}

// Update modifies an existing template
func (r *TemplateRepository) Update(ctx context.Context, template db.Template) (*db.Template, error) {
	query := `
		UPDATE templates SET
			name = $2,
			description = $3,
			preview_image_url = $4,
			config = $5,
			is_default = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, name, description, preview_image_url, config, is_default, created_at, updated_at
	`

	var updatedTemplate db.Template
	err := r.db.QueryRowxContext(ctx, query,
		template.ID,
		template.Name,
		db.TextToString(template.Description),
		db.TextToString(template.PreviewImageUrl),
		template.Config,
		template.IsDefault,
	).StructScan(&updatedTemplate)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return &updatedTemplate, nil
}

// Delete removes a template by ID
func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM templates WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTemplateNotFound
	}

	return nil
}
