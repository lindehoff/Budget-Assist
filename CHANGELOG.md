# Budget-Assist Changelog

## <small>2.6.1 (2025-02-26)</small>

* **design:** update CLI design for runtime user insights ([](https://github.com/lindehoff/Budget-Assist/commit/b8885b74fff941d6549a03c3a30bcc1b1703f3dd))
  • Update System Architecture to include runtime user insights in process command
• Update AI Integration design to handle runtime insights during processing
• Update Implementation Plan to focus on runtime insights integration
• Remove stored user insights from prompt management

Changes:
- Add ProcessOptions with runtime insight fields
- Update AI service interface for runtime insights
- Add document processing pipeline flow
- Add example CLI usage with insights
- Update implementation tasks and timeline

## 2.6.0 (2025-02-26)

* Enhance document processing capabilities 📄✨ ([](https://github.com/lindehoff/Budget-Assist/commit/76d0e83c4548fe2d06fe963eb9c2a34f007fdc6d))
  - Add support for processing bank statements and CSV files.
- Implement prompts for bank statement extraction and CSV transaction analysis.
- Update CLI commands to include processing for bank statements and CSV files.
- Improve transaction categorization prompts for clarity and structure.
- Ensure successful integration of AI services for transaction analysis.

## 2.5.0 (2025-02-25)

* Enhance OpenAI integration and document processing capabilities 🎉 ([](https://github.com/lindehoff/Budget-Assist/commit/5ec5c3184b8eabba2ed376bab20e59ea6c7eec58))
  - Updated the overview to reflect the new functionalities of the system,
  emphasizing PDF document processing and transaction categorization. 📄
- Introduced new core components, including the DocumentData and
  ExtractionResult types for better document handling. 🛠️
- Added methods for adding examples to prompts and improved the
  AnalysisRequest structure by removing unnecessary fields. ✨
- Updated prompt templates for bank statements, CSV transactions,
  invoices, and receipts to enhance extraction accuracy. 📊
- Removed deprecated types and methods related to document data
  extraction and processing. 🗑️
- Improved the CLI integration by adding new commands for processing
  and categorizing transactions. ⚙️
- Enhanced configuration structures for OpenAI and PDF processing
  to support new features and optimizations. 🔧
- Refined error handling and monitoring strategies to ensure
  robustness in document processing and API interactions. 🔍

## <small>2.4.1 (2025-02-25)</small>

* **category:** resolve race conditions in concurrent category operations ([](https://github.com/lindehoff/Budget-Assist/commit/d150c4dff145d3d7786128f8e66ba6eb7f397e3f))
  • Added proper synchronization using sync.RWMutex for concurrent category operations
• Reduced concurrent operations from 10 to 5 for better test stability
• Protected shared state access with appropriate locks in TestConcurrent_category_operations
• Implemented retries for category updates to handle race conditions gracefully
• Verified fix with race detector enabled (-race flag)

Test coverage improvements:
- internal/api: 100%
- internal/category: 91.7%
- internal/core: 100%
- internal/db: 52.8%
- internal/docprocess: 84.4%
- internal/processor: 93.8%

## 2.4.0 (2025-02-25)

* **mockStore:** add concurrency support with RWMutex 🛠️ ([](https://github.com/lindehoff/Budget-Assist/commit/5d6b1fc8b9cbc9710ed204c212d9729c5e8a62b6))
  - Introduced a RWMutex to the mockStore struct to ensure safe concurrent access
  to categories and translations. 🔒
- Updated CreateCategory, UpdateCategory, GetCategoryByID, ListCategories,
  CreateTranslation, DeleteCategory, GetCategoryTypeByID, GetPromptByType,
  UpdatePrompt, and ListPrompts methods to use locking mechanisms. 🔄
- This change enhances the mockStore's ability to handle concurrent operations
  in tests, improving reliability and preventing race conditions. 🚀

## 2.3.0 (2025-02-24)

* **cmd:** ✨ implement category management commands ([](https://github.com/lindehoff/Budget-Assist/commit/a60cb982d7e20c030a5b492d455d32a0afac02c2))
  Adds CLI commands for managing transaction categories with support for:
• 📋 List categories with table/JSON output formats
• ➕ Add new categories with name and description
• 🔄 Update existing categories (name, description, active status)
• 🗑️ Soft delete categories with confirmation prompt
• 🛡️ Custom error handling for category operations
• 🔧 Utility functions for DB store, AI service, and output formatting

Note: Parent category support and budget/color features are marked as TODO 📝
* **db:** 🗃️  Add Prompt model and database migrations ([](https://github.com/lindehoff/Budget-Assist/commit/c47ac5037e85ef1b9f6951d2858ae6ad9b322daa))
  - Add new Prompt model for storing AI prompt templates
- Add database migrations for Prompt table
- Implement CRUD operations in SQLStore interface
- Add robust error handling for database operations

refactor(ai): 🔄 Migrate prompt management to use database storage

- Convert in-memory prompt storage to database-backed system
- Enhance version increment logic (major.minor.patch)
- Add mutex locks for concurrent access safety
- Improve error handling and validation

test(ai): 🧪 Enhance prompt management test coverage

- Add comprehensive test cases for database operations
- Add concurrent operation tests
- Improve test error messages with detailed comparisons
- Update mock implementations for database testing

docs(ai): 📝 Update code documentation

- Add detailed comments for new database methods
- Update function documentation to reflect database usage
- Add usage examples in comments
- Improve error message clarity
* **ai:** improve prompt template organization and fix tests 🔨 ([](https://github.com/lindehoff/Budget-Assist/commit/5123e199d207e57a9446d0c59d6c1413d4e3e468))
  * 🏗️ Add new prompt types for better separation of concerns:
  - TransactionAnalysisPrompt
  - CategorySuggestionPrompt
  - Keep existing DocumentExtractionPrompt

* 📝 Update template data structures to match their specific use cases:
  - Analysis template: Description, Amount, Date fields
  - Suggestion template: Description, Categories fields
  - Document template: Content field

* ✅ Fix template execution errors in tests by using correct prompt types
* 🧪 Add new prompt_manager with comprehensive test coverage
* 📚 Update implementation documentation in Iteration3-Tasks.md

## 2.2.0 (2025-02-24)

* 🤖 refactor(ai): update design for OpenAI integration ([](https://github.com/lindehoff/Budget-Assist/commit/a5a3c5ac15b1f823d342d6bd082f65d91b4a2195))
  🔄 Core Changes:
- Replace custom AI implementation with OpenAI service integration
- Update PDF processing pipeline to use pdfcpu and OpenAI
- Add document-specific prompts for different document types
- Enhance category management with hierarchical paths

🛠️ Technical Improvements:
- Add CLI commands for prompt and category management
- Implement proper error handling and validation
- Add cost tracking and monitoring capabilities

📝 The changes reflect our decision to use OpenAI's services instead of
implementing our own model, while maintaining the flexibility to handle
different document types and categories.

🔧 Technical Details:
- Update AIService interface to include prompt management
- Add document-specific prompt templates
- Enhance PDF processing pipeline with OpenAI integration
- Add validation rules and confidence scoring
- Implement proper error handling for OpenAI responses
- Add cost tracking and rate limiting

✨ This update streamlines our AI integration while improving
maintainability and extensibility.
* Merge pull request #21 from lindehoff/feature/openai-integration-design ([](https://github.com/lindehoff/Budget-Assist/commit/9d268e560a383fc204fb0dc6a1b356d3670759c7))
  🤖 refactor(ai): update design for OpenAI integration
* Merge pull request #22 from lindehoff/feat/category-management ([](https://github.com/lindehoff/Budget-Assist/commit/fd5fa94687c76b9a256bddd7471a8801e1b37bfc))
  feat(category): implement category management system
* **category:** implement category management system ([](https://github.com/lindehoff/Budget-Assist/commit/255b9394ae799531c0cd25b80ebb3846f83eb31b))
  - ✨ Add category manager with CRUD operations
- 🧪 Add comprehensive test suite for category management
- 📦 Create store interface for database operations
- 🔄 Implement translation support

test(ai): enhance prompt testing

- 🧪 Update prompt test cases
- ✅ Add validation scenarios
- 🔄 Improve test coverage

docs(plan): update implementation documentation

- 📝 Update Implementation Plan with completed features
- ✨ Add Strategic Decisions section
- 🎯 Update Iteration 3 tasks and status
- ✅ Mark completed category management tasks

build(deps): clean up dependencies

- 🧹 Remove testify dependency
- 📦 Update go.mod and go.sum
- 🔨 Enhance error handling in db package
* **docprocess:** improve test coverage to 84.4% ([](https://github.com/lindehoff/Budget-Assist/commit/50da57ba4cb2dd225e96f2b94dc0e14b46097acd))
  * 🧪 Add comprehensive tests for error types and methods in types.go (100% coverage)
* 🔍 Add test helper functions for creating valid and corrupted PDF files
* ✅ Add test cases for PDF validation and processing (85.7% and 77.8% coverage)
* 🔄 Implement integration test with known PDF content
* 🐛 Fix error unwrapping test to correctly handle error comparisons

## 2.1.0 (2025-02-24)

* Merge pull request #20 from lindehoff/feature/ai-service-integration ([](https://github.com/lindehoff/Budget-Assist/commit/4154f1f5928e2da9aca051024fa5428264ceb54a))
  Feature/ai service integration
* **ai:** implement OpenAI service integration 🤖 ([](https://github.com/lindehoff/Budget-Assist/commit/bfbf6a9f1967cc8a62fe25a49e6fd09b1fb4654e))
  • 🎯 Added core AI service interface and OpenAI implementation
• 📝 Created prompt template system with example-based learning
• ⚡ Implemented rate limiting and retry logic with exponential backoff
• 🛡️ Added comprehensive error handling and monitoring
• 📦 Updated dependencies to include rate limiting support
• ✅ Updated iteration 3 task list with completed items
* **lint:** adjust golangci-lint configuration for better test handling ([](https://github.com/lindehoff/Budget-Assist/commit/8bc63d239bdb48c9c5923533bffbecfb413fb7d7))
  • Allow test files to remain in the same package as the code they test
• Configure more permissive linting rules for test files:
  - Disable field alignment checks (govet)
  - Allow longer lines and higher complexity (lll, gocognit)
  - Skip whitespace and style rules (wsl, revive)
  - Allow TODO comments and context usage (godox, noctx)
• Disable opinionated linters for all files:
  - Comment formatting (godot)
  - TODO comments (godox)
  - Style checks (revive)
  - Formatting (gofumpt)
  - Nolint directives (nolintlint)

This change improves test maintainability while keeping strict standards for production code.

## <small>2.0.3 (2025-02-24)</small>

* **release:** 2.0.3 [skip ci] ([](https://github.com/lindehoff/Budget-Assist/commit/2ba03c9e6c5e461a9d398e9708d0684d11e42dfc))
  ## <small>2.0.3 (2025-02-24)</small>

* **ci:** improve semantic-release version capture and handling ([](https://github.com/lindehoff/Budget-Assist/commit/ade400b9a47742330d9d0a3018cd2240c30eff39))
  * Add dry-run step to extract next version before actual release
* Store version in GitHub output for subsequent steps
* Update conditional checks to use captured version
* Fix version reference in upload-release-action
* **ci:** improve semantic-release version capture and handling ([](https://github.com/lindehoff/Budget-Assist/commit/ade400b9a47742330d9d0a3018cd2240c30eff39))
  * Add dry-run step to extract next version before actual release
* Store version in GitHub output for subsequent steps
* Update conditional checks to use captured version
* Fix version reference in upload-release-action
* **ci:** switch to official semantic-release action ([](https://github.com/lindehoff/Budget-Assist/commit/7c309126eb08549dd89c7258ff56d344dd710848))
  * Replace custom semantic-release implementation with cycjimmy/semantic-release-action
* Update checkout action configuration for better Git credentials handling
* Fix version tag reference to use semantic-release output variables
* Remove manual Git configuration as it's handled by the action

## <small>2.0.3 (2025-02-24)</small>

* **ci:** improve semantic-release version capture and handling ([](https://github.com/lindehoff/Budget-Assist/commit/ade400b9a47742330d9d0a3018cd2240c30eff39))
  * Add dry-run step to extract next version before actual release
* Store version in GitHub output for subsequent steps
* Update conditional checks to use captured version
* Fix version reference in upload-release-action

## <small>2.0.2 (2025-02-24)</small>

* **ci:** prevent duplicate releases by using semantic version tag ([](https://github.com/lindehoff/Budget-Assist/commit/2daea4fa8cb2a185bc4c1fea69ade947816591f0))
  * Update upload-release-action to use semantic-release version output instead of github.ref
* Add 'v' prefix to match semantic-release tag format

## <small>2.0.1 (2025-02-24)</small>

* **ci:** update release workflow to use semantic versioning correctly ([](https://github.com/lindehoff/Budget-Assist/commit/505304f552e621c43ccb894ac72e95f36572ee9c))
  * Add Git configuration for GitHub Actions bot authentication
* Add semantic-release step ID for output tracking
* Add conditional steps based on semantic-release success
* Update tag reference to use github.ref for consistent versioning

## 2.0.0 (2025-02-24)

* Merge pull request #18 from lindehoff/docs/update-iteration2-progress ([](https://github.com/lindehoff/Budget-Assist/commit/7d683397ecb7c383b1ec37221b34c3a1962e40f3))
* Merge pull request #19 from lindehoff/feat/csv-processor-seb ([](https://github.com/lindehoff/Budget-Assist/commit/b8560d9b940154eb9d8b58f09a42934fdc85b7d1))
  feat(processor): implement SEB CSV processor and complete Iteration 2
* **docprocess:** 📄 implement document processing core and PDF handler 🛠️ ([](https://github.com/lindehoff/Budget-Assist/commit/2e767299af9df703fb5dfc72279a1c21bb6f49ea))
  - ✨ Add document processor interface and factory pattern for extensibility
- 📝 Implement PDF processing with pdfcpu for text extraction
- 🛡️ Create robust error handling system with processing stages
- 🧪 Add comprehensive test suite with table-driven tests
- 🏗️ Implement transaction model with proper field alignment

Technical details:
- 🔍 Add ProcessingError with stage-based error handling
- 🎯 Implement DocumentProcessor and ProcessorFactory interfaces
- 📋 Add PDF text extraction with proper temp file handling
- ✅ Create test framework ready for integration tests
- 📚 Update iteration 2 documentation to reflect implementation
  Breaking changes: none 🚀
* **processor:** implement SEB CSV processor and complete Iteration 2 ([](https://github.com/lindehoff/Budget-Assist/commit/6f8fedd24978b925933029e07a2604c9917e4fa1))
  🎯 Core Features:
- Implemented SEB CSV processor with comprehensive error handling
- Added table-driven tests following Go standards
- Updated PDF processor imports to use correct pdfcpu paths
- Optimized struct field alignment for better memory usage

📝 Documentation:
- Marked Iteration 2 as complete
- Moved non-critical tasks to Future Improvements
- Updated success criteria and review checklist
- Added detailed implementation notes

🧪 Testing:
- Added comprehensive test suite for SEB processor
- Included success and error test cases
- Added IO failure scenarios
- Implemented proper validation blocks

🔧 Technical Improvements:
- Fixed dependency management in go.mod
- Optimized memory layout of Transaction and ProcessingError structs
- Added structured logging with proper context
- Implemented early returns for validation
* **ci:** resolve semantic-release versioning and Git permission issues ([](https://github.com/lindehoff/Budget-Assist/commit/f0ec774728ee33189f84b4780808a235680db6e3))
  * Reorder workflow steps to run semantic-release before building binaries
* Add proper Git configuration for GitHub Actions bot
* Update Makefile version detection to use semantic version tags
* Fix release artifact upload to use correct semantic version tag
* Add conditional steps to ensure proper release flow
* **ci:** resolve version handling in release workflow and Makefile ([](https://github.com/lindehoff/Budget-Assist/commit/96b47792532f34698ce6cba189ec2acf368b47fc))
  * Update VERSION variable in Makefile to use consistent tag detection
* Add explicit release tag retrieval step in GitHub Actions workflow
* Update upload-release-action to use correct tag reference
* Remove conditional release step to prevent race conditions


### Breaking changes

* none 🚀

## 1.7.0 (2025-02-23)

* Merge pull request #17 from lindehoff/feat/iteration-1-complete ([](https://github.com/lindehoff/Budget-Assist/commit/7b9c6a606bf2e8e4c5b99abce9be19f152f555cd))
* improve documentation organization and accessibility ([](https://github.com/lindehoff/Budget-Assist/commit/ecf0b178ad01c823598fe1cc7f84f4c06d4cfcda))
  - Restructure documentation section in README with clear categories for users and developers
- Add quick start guide with installation and configuration examples
- Create comprehensive guides for installation, configuration, and CLI usage
- Add contributing guidelines with detailed instructions for developers
- Improve help section with multiple support channels
- Update documentation links to point to new guide locations

The documentation is now better organized and more accessible, making it easier for both users
and developers to find the information they need.
* **core:** complete iteration 1 core infrastructure ([](https://github.com/lindehoff/Budget-Assist/commit/e616451246c3044968ea2aab3b738c3d9762ab6a))
  Complete foundational project setup and infrastructure implementation:

• ✅ Project structure and CLI framework
• 🗄️ Database setup with GORM and migrations
• 🔄 CI/CD pipeline with GitHub Actions
• 📚 Comprehensive documentation
• ✨ Error handling implementation
• 🧪 Test coverage and reporting

All review checklist items completed and verified against design requirements.
  Closes #1
* **ci:** improve semantic-release configuration ([](https://github.com/lindehoff/Budget-Assist/commit/002898e85b83e347a1981da355d573310621dd0e))
  - Add dedicated .releaserc.json file
- Update GitHub Actions workflow permissions
- Fix token handling in checkout step
- Clean up Git references# Please enter the commit message for your changes. Lines starting

## 1.6.0 (2025-02-23)

* ✨ feat(cicd): enhance CI/CD pipeline and code quality - Configure CI workflow with Go 1.24.0, set up linting with golangci-lint and ESLint, configure test running with race detection, add multi-platform build process ([](https://github.com/lindehoff/Budget-Assist/commit/f9f1a3590054f3fa0ccb8bbf9324c27f6b2bbafb))
* Merge pull request #16 from lindehoff/feat/enhance-cicd-pipeline ([](https://github.com/lindehoff/Budget-Assist/commit/1497a5a320f46e5e58c018ed758d397859bf38ff))
  Feat/enhance cicd pipeline
* configure codecov token for coverage uploads ([](https://github.com/lindehoff/Budget-Assist/commit/4aeb7fa3f8af326d7ab3ae80e864ab528a6b6f65))
  - Add CODECOV_TOKEN secret configuration to codecov-action
- Enable authenticated coverage report uploads
- Fix rate limiting issues with Codecov uploads

Note: Requires CODECOV_TOKEN to be configured in repository secrets
* **ci:** 🚀 enhance CI/CD pipeline and code quality ([](https://github.com/lindehoff/Budget-Assist/commit/7b4a84a1e4f905aee0090fc5af31fbb84e2c1671))
  * 🔧 Add CI workflow with Go and Node.js setup\n* 🔍 Configure golangci-lint with custom rules\n* ✨ Set up ESLint for TypeScript/JavaScript\n* 🧪 Add test running with race detection and coverage reporting\n* 📦 Implement multi-platform build process with checksums\n* 🏷️ Configure semantic release with automated versioning\n* ⬆️ Update Go version to 1.24.0\n* 🗃️ Optimize struct field alignment in database models\n* 🛡️ Add comprehensive error handling\n* 📝 Update iteration 1 tasks documentation\n* 🙈 Update .gitignore for better file management
* **ci:** 🚀 enhance CI/CD pipeline and code quality ([](https://github.com/lindehoff/Budget-Assist/commit/7d635f432badac9a6d5995551af499fdc7edee08))
  * 🔧 Add CI workflow with Go and Node.js setup\n* 🔍 Configure golangci-lint with custom rules\n* ✨ Set up ESLint for TypeScript/JavaScript\n* 🧪 Add test running with race detection and coverage reporting\n* 📦 Implement multi-platform build process with checksums\n* 🏷️ Configure semantic release with automated versioning\n* ⬆️ Update Go version to 1.24.0\n* 🗃️ Optimize struct field alignment in database models\n* 🛡️ Add comprehensive error handling\n* 📝 Update iteration 1 tasks documentation
* **web:** initialize TypeScript and React setup ([](https://github.com/lindehoff/Budget-Assist/commit/f8bbe56003743320d4abc85d1ee97eb1f83af66a))
  - Add basic TypeScript configuration and types
- Create initial React component structure
- Configure ESLint for TypeScript
* **db:** add foreign key constraints and currency validation ([](https://github.com/lindehoff/Budget-Assist/commit/226cbfc7c3c702aac86171ec89124c2c7a5adfcc))
  - Add explicit foreign key constraints to CategoryType, Transaction and Budget models
- Implement BeforeCreate hook for Transaction to validate currency
- Add coverage.txt to gitignore to prevent tracking test coverage files
- Fix seed test by adding missing SeedPredefinedCategories call

## <small>1.5.1 (2025-02-23)</small>

* Merge pull request #12 from lindehoff/chore/fix-semantic-release-config ([](https://github.com/lindehoff/Budget-Assist/commit/99d9a3637c68a720d48dbbc7438b87ad22a8641d))
  fix(release): consolidate semantic-release configuration 📝
* Merge pull request #13 from lindehoff/chore/fix-semantic-release-config ([](https://github.com/lindehoff/Budget-Assist/commit/b8366d59f9668d318d4681ae40d0c560ded111a8))
  Chore/fix semantic release config
* Merge pull request #14 from lindehoff/chore/fix-semantic-release-config ([](https://github.com/lindehoff/Budget-Assist/commit/933f23d880fbb04b4d4d3037a27cef0793b6175f))
  fix(release): resolve date handling issue in release notes 🐛
* Merge pull request #15 from lindehoff/chore/fix-semantic-release-config ([](https://github.com/lindehoff/Budget-Assist/commit/c3e768ca0ac51f2fc51bb90a401dd00f7d095048))
  fix(ci): handle missing package-lock.json 🔧
* add package-lock.json [skip ci] ([](https://github.com/lindehoff/Budget-Assist/commit/cc85bc18ddffcd4b68c261c07680f74882cb813b))
* **ci:** handle missing package-lock.json 🔧 ([](https://github.com/lindehoff/Budget-Assist/commit/98a3529d414e00ae4ac84070230735fa05467467))
  Updated release workflow to handle missing lock file:

- Generate package-lock.json during CI

- Configure git user for lock file commit

- Temporarily disable npm cache until lock file exists

- Add fallback for when lock file already exists
* **ci:** update release workflow to use local dependencies 🔧 ([](https://github.com/lindehoff/Budget-Assist/commit/616a8c0d1e1b80f11b88abfcb47ec112d67aca68))
  Updated the release workflow configuration:

- Removed global package installations

- Using local dependencies from package.json

- Added npm ci for deterministic installs

- Removed unnecessary conventional-changelog-angular

This change ensures we use the correct conventional-changelog-conventionalcommits package and follows better practices for dependency management in CI.
* **release:** add conventional-changelog-conventionalcommits dependency 🔧 ([](https://github.com/lindehoff/Budget-Assist/commit/271fd18e385653bcca4787e710a897593437d21c))
  Added missing dependency required for semantic-release:

- Added conventional-changelog-conventionalcommits package

- This package is required for parsing conventional commits

- Fixes the MODULE_NOT_FOUND error in the release workflow
* **release:** consolidate semantic-release configuration 📝 ([](https://github.com/lindehoff/Budget-Assist/commit/67cc832989c28059ca4342d6f2c52ef46b2abace))
  Moved semantic-release configuration from .releaserc.json to package.json:

- Consolidated all release configuration in one place

- Added proper changelog title

- Configured commit body inclusion in changelog

- Set up proper commit message parsing

This change ensures that all semantic-release configuration is properly recognized and applied, fixing issues with changelog generation and commit message formatting.
* **release:** resolve date handling issue in release notes 🐛 ([](https://github.com/lindehoff/Budget-Assist/commit/86fd7adf5126bf4bb15df5c08249629ad02f8db7))
  Fixed release notes generation issues:

- Disabled committer date in release notes to avoid date parsing errors

- Simplified commit transformation configuration

- Set specific Node.js version (20.11.0) for better compatibility

- Added npm cache for faster CI runs

# [1.5.0](https://github.com/lindehoff/Budget-Assist/compare/v1.4.0...v1.5.0) (2025-02-23)


### Features

* **cli:** add quiet mode to version command 🤫 ([c398f12](https://github.com/lindehoff/Budget-Assist/commit/c398f1221fb52c12d179b0a74de9ca3155f6d1c8))

# [1.4.0](https://github.com/lindehoff/Budget-Assist/compare/v1.3.0...v1.4.0) (2025-02-23)


### Features

* **cli:** add short version output option 🔍 ([e7212dd](https://github.com/lindehoff/Budget-Assist/commit/e7212dd52d4f3b8da9f23f076aa47f36e8b3a105))

# [1.3.0](https://github.com/lindehoff/Budget-Assist/compare/v1.2.0...v1.3.0) (2025-02-23)


### Features

* **cli:** add JSON output format to version command 🔄 ([a1b74ea](https://github.com/lindehoff/Budget-Assist/commit/a1b74eaacf9f5d2ffe2d4e596287206f382d6972))

# [1.2.0](https://github.com/lindehoff/Budget-Assist/compare/v1.1.0...v1.2.0) (2025-02-23)


### Features

* **build:** implement proper version handling with build information 🏗️ ([4fb957e](https://github.com/lindehoff/Budget-Assist/commit/4fb957ea8deef9ba5ea48c20f0730347f6cd8dce))

# [1.1.0](https://github.com/lindehoff/Budget-Assist/compare/v1.0.0...v1.1.0) (2025-02-23)


### Features

* **db:** implement core database layer with tests 🎉 ([98babe9](https://github.com/lindehoff/Budget-Assist/commit/98babe9461cd86447c832209f68b5c516b925e69))

# 1.0.0 (2025-02-23)


### Features

* **cli:** implement core CLI infrastructure and basic commands ([1349cca](https://github.com/lindehoff/Budget-Assist/commit/1349ccaf148e66187f861600c17937d4f45bd3ed))
