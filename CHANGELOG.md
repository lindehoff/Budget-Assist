# Budget-Assist Changelog

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
