package db

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSubcategory_TagOperations(t *testing.T) {
	store, _ := createTestStore(t)
	ctx := context.Background()

	// Create a test subcategory with tags
	tags := []string{"food", "groceries", "essential"}
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		t.Fatalf("Failed to marshal tags: %v", err)
	}

	// Create a translation for the subcategory
	translation := Translation{
		EntityType:   string(EntityTypeSubcategory),
		LanguageCode: LangEN,
		Name:         "Test Subcategory",
		Description:  "Test Description",
	}

	subcategory := &Subcategory{
		Name:           "Test Subcategory",
		Description:    "Test Description",
		CategoryTypeID: 1,
		IsActive:       true,
		Tags:           string(tagsJSON),
		Translations:   []Translation{translation},
	}

	// Test creating subcategory with tags
	err = store.CreateSubcategory(ctx, subcategory)
	if err != nil {
		t.Fatalf("Failed to create subcategory: %v", err)
	}

	// Test GetTags
	t.Run("GetTags", func(t *testing.T) {
		retrievedTags := subcategory.GetTags()
		if len(retrievedTags) != len(tags) {
			t.Errorf("GetTags() returned %d tags, want %d", len(retrievedTags), len(tags))
		}
		for i, tag := range tags {
			if retrievedTags[i] != tag {
				t.Errorf("GetTags()[%d] = %s, want %s", i, retrievedTags[i], tag)
			}
		}
	})

	// Test HasTag
	t.Run("HasTag", func(t *testing.T) {
		if !subcategory.HasTag("food") {
			t.Error("HasTag('food') = false, want true")
		}
		if subcategory.HasTag("nonexistent") {
			t.Error("HasTag('nonexistent') = true, want false")
		}
	})

	// Test FindSubcategoriesByTag
	t.Run("FindSubcategoriesByTag", func(t *testing.T) {
		// Test finding existing tag
		found, err := store.FindSubcategoriesByTag(ctx, "food")
		if err != nil {
			t.Fatalf("FindSubcategoriesByTag() error = %v", err)
		}
		if len(found) != 1 {
			t.Errorf("FindSubcategoriesByTag() returned %d subcategories, want 1", len(found))
		}
		if len(found) > 0 && found[0].ID != subcategory.ID {
			t.Errorf("FindSubcategoriesByTag() returned subcategory ID %d, want %d", found[0].ID, subcategory.ID)
		}

		// Test finding non-existent tag
		found, err = store.FindSubcategoriesByTag(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("FindSubcategoriesByTag() error = %v", err)
		}
		if len(found) != 0 {
			t.Errorf("FindSubcategoriesByTag() returned %d subcategories for non-existent tag, want 0", len(found))
		}
	})

	// Test updating tags
	t.Run("UpdateTags", func(t *testing.T) {
		newTags := []string{"food", "groceries", "essential", "monthly"}
		newTagsJSON, err := json.Marshal(newTags)
		if err != nil {
			t.Fatalf("Failed to marshal new tags: %v", err)
		}

		subcategory.Tags = string(newTagsJSON)
		err = store.UpdateSubcategory(ctx, subcategory)
		if err != nil {
			t.Fatalf("Failed to update subcategory tags: %v", err)
		}

		// Verify the update
		updated, err := store.GetSubcategoryByID(ctx, subcategory.ID)
		if err != nil {
			t.Fatalf("Failed to get updated subcategory: %v", err)
		}

		updatedTags := updated.GetTags()
		if len(updatedTags) != len(newTags) {
			t.Errorf("Updated subcategory has %d tags, want %d", len(updatedTags), len(newTags))
		}
		for i, tag := range newTags {
			if updatedTags[i] != tag {
				t.Errorf("Updated tags[%d] = %s, want %s", i, updatedTags[i], tag)
			}
		}
	})
}

func TestSubcategory_EmptyTags(t *testing.T) {
	subcategory := &Subcategory{
		Name:           "Test Subcategory",
		Description:    "Test Description",
		CategoryTypeID: 1,
		IsActive:       true,
	}

	// Test GetTags with empty tags
	t.Run("GetTags_Empty", func(t *testing.T) {
		tags := subcategory.GetTags()
		if len(tags) != 0 {
			t.Errorf("GetTags() with empty tags returned %d tags, want 0", len(tags))
		}
	})

	// Test HasTag with empty tags
	t.Run("HasTag_Empty", func(t *testing.T) {
		if subcategory.HasTag("any") {
			t.Error("HasTag() with empty tags returned true, want false")
		}
	})
}

func TestSubcategory_InvalidTags(t *testing.T) {
	subcategory := &Subcategory{
		Name:           "Test Subcategory",
		Description:    "Test Description",
		CategoryTypeID: 1,
		IsActive:       true,
		Tags:           "invalid json",
	}

	// Test GetTags with invalid JSON
	t.Run("GetTags_InvalidJSON", func(t *testing.T) {
		tags := subcategory.GetTags()
		if len(tags) != 0 {
			t.Errorf("GetTags() with invalid JSON returned %d tags, want 0", len(tags))
		}
	})

	// Test HasTag with invalid JSON
	t.Run("HasTag_InvalidJSON", func(t *testing.T) {
		if subcategory.HasTag("any") {
			t.Error("HasTag() with invalid JSON returned true, want false")
		}
	})
}
