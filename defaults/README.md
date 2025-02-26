# Default Prompts and Categories

This directory contains default prompts and categories for Budget-Assist. These files can be imported when setting up a new instance of the application.

## Files

- `prompts.json`: Contains default prompts for analyzing bills, receipts, bank statements, and categorizing transactions. Each prompt includes translations in English and Swedish.
- `categories.json`: Contains default categories and subcategories for organizing transactions. Each category includes translations in English and Swedish.

## Usage

To import the default prompts and categories, use the following commands:

```bash
# Import default prompts
budget-assist prompt import --file defaults/prompts.json

# Import default categories
budget-assist category import --file defaults/categories.json
```

## Structure

### Prompts

The `prompts.json` file contains:
- Bill analysis prompts
- Receipt analysis prompts
- Bank statement analysis prompts
- Transaction categorization prompts

Each prompt includes:
- Type identifier
- Base prompt template
- System prompt
- User prompt
- Translations (English and Swedish)

### Categories

The `categories.json` file contains a hierarchical structure of:
- Main categories (e.g., Income, Housing, Vehicle)
- Subcategories for each main category
- Translations for each category and subcategory (English and Swedish)
- Descriptions for better understanding of each category

## Customization

You can customize these files by:
1. Exporting your current configuration
2. Modifying the exported JSON files
3. Importing the modified files

Example:
```bash
# Export current prompts
budget-assist prompt export --file my-prompts.json

# Export current categories
budget-assist category export --file my-categories.json

# After modification, import your custom files
budget-assist prompt import --file my-prompts.json
budget-assist category import --file my-categories.json
```

## Adding New Languages

To add support for a new language:
1. Add a new language code in the `translations` object for each prompt and category
2. Provide translated versions of:
   - Names
   - Descriptions
   - Prompts (for prompt templates)

Example adding German translations:
```json
{
  "translations": {
    "en": { ... },
    "sv": { ... },
    "de": {
      "name": "German name",
      "description": "German description"
    }
  }
}
``` 