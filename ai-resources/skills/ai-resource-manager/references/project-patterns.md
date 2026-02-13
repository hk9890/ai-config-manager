# Project Pattern → Skill Mappings

## Overview

This reference guides skill recommendations based on project analysis. It provides pattern matching rules for the auto-discovery workflow to quickly identify project technologies and recommend relevant skills.

**Key Principles:**
- **Quick scan approach**: Check only common indicator files, don't read entire codebase
- **Root-level priority**: Files in project root are strongest indicators
- **Multiple matches**: Projects often match multiple patterns simultaneously
- **Progressive refinement**: Start broad, then refine based on additional indicators

## Node.js Projects

**Detection patterns:**
- `package.json` present in project root
- `.js`, `.ts`, `.mjs`, `.cjs` files in `src/` or root
- `node_modules/` directory (less reliable, may be gitignored)
- `yarn.lock` or `package-lock.json` for package manager detection

**Recommended skills:**
- **npm-helper** - Package management, dependency installation
- **jest-testing** - If `jest.config.js`, `jest.config.ts`, or `"jest"` section in package.json
- **playwright-e2e** - If `playwright.config.ts` or `playwright.config.js` present
- **typescript-helper** - If `tsconfig.json` present or `.ts` files detected
- **webpack-bundler** - If `webpack.config.js` present
- **vite-bundler** - If `vite.config.ts` or `vite.config.js` present

**Framework-specific patterns:**
- **react-helper** - If `package.json` includes `react` dependency
- **vue-helper** - If `package.json` includes `vue` dependency
- **next-helper** - If `next.config.js` present or `next` in dependencies
- **express-helper** - If `package.json` includes `express` dependency

## Python Projects

**Detection patterns:**
- `requirements.txt` in project root
- `pyproject.toml` in project root (modern Python projects)
- `setup.py` or `setup.cfg` present
- `.py` files in `src/`, `lib/`, or root
- `Pipfile` and `Pipfile.lock` (pipenv projects)
- `poetry.lock` (poetry projects)

**Recommended skills:**
- **python-helper** - General Python development support
- **pytest-testing** - If `pytest.ini`, `pyproject.toml` with `[tool.pytest]`, or `tests/` directory with `test_*.py` files
- **poetry-helper** - If `pyproject.toml` and `poetry.lock` present
- **pipenv-helper** - If `Pipfile` present
- **django-helper** - If `manage.py` present or `django` in requirements
- **flask-helper** - If `flask` in requirements.txt
- **fastapi-helper** - If `fastapi` in requirements.txt

**Additional indicators:**
- `requirements-dev.txt` or `requirements-test.txt` - separate test dependencies
- `setup.py` with `install_requires` - package distribution
- `tox.ini` - multi-environment testing

## Docker Projects

**Detection patterns:**
- `Dockerfile` in project root or subdirectories
- `docker-compose.yml` or `docker-compose.yaml` present
- `.dockerignore` file present
- `docker/` directory containing Dockerfiles

**Recommended skills:**
- **docker-helper** - Container building, running, management
- **docker-compose-helper** - If `docker-compose.yml` present
- **kubernetes-helper** - If `k8s/`, `kubernetes/`, or `.yaml` files with `kind:` fields present

**Multi-stage patterns:**
- Check Dockerfile for `FROM ... AS` statements (multi-stage builds)
- Check for specific base images (node, python, nginx, alpine)

## CI/CD Projects

**Detection patterns:**
- `.github/workflows/` directory (GitHub Actions)
- `.gitlab-ci.yml` file (GitLab CI)
- `Jenkinsfile` in project root (Jenkins)
- `.circleci/config.yml` (CircleCI)
- `.travis.yml` (Travis CI)
- `azure-pipelines.yml` (Azure DevOps)
- `.drone.yml` (Drone CI)

**Recommended skills:**
- **github-actions** - For `.github/workflows/` directory
- **gitlab-ci-helper** - For `.gitlab-ci.yml`
- **jenkins-helper** - For `Jenkinsfile`
- **ci-helper** - General CI/CD support for other platforms
- **deployment-helper** - If deployment steps detected in CI config

**Workflow analysis:**
- Check for test, build, deploy stages
- Identify deployment targets (AWS, GCP, Azure, Heroku)
- Look for artifact publishing (npm, PyPI, Docker Hub)

## Testing Projects

**Detection patterns:**
- `tests/` or `test/` directory in project root
- `__tests__/` directory (JavaScript convention)
- Test files: `*.test.js`, `*.spec.js`, `*_test.py`, `test_*.py`
- `cypress/` directory (Cypress E2E tests)
- `e2e/` or `integration/` directories

**Recommended skills:**
- **jest-testing** - If Jest config or `.test.js`/`.spec.js` files
- **pytest-testing** - If pytest config or `test_*.py` files
- **cypress-testing** - If `cypress.json` or `cypress/` directory
- **playwright-e2e** - If `playwright.config.ts` present
- **selenium-testing** - If selenium dependencies detected
- **unittest-helper** - For Python `unittest` framework

**Test type indicators:**
- `tests/unit/` - unit testing skills
- `tests/integration/` - integration testing skills
- `tests/e2e/` - end-to-end testing skills
- `tests/performance/` - performance testing skills

## Documentation Projects

**Detection patterns:**
- `docs/` directory with multiple `.md` files
- `README.md` extensive (>200 lines)
- `.md` files abundant (>10 markdown files)
- `mkdocs.yml` (MkDocs documentation)
- `Sphinx/` or `conf.py` (Sphinx documentation)
- `docusaurus.config.js` (Docusaurus)
- `.readthedocs.yml` (Read the Docs)

**Recommended skills:**
- **documentation-helper** - General documentation support
- **mkdocs-helper** - If `mkdocs.yml` present
- **sphinx-helper** - If `conf.py` and `Sphinx/` present
- **docusaurus-helper** - If `docusaurus.config.js` present
- **markdown-helper** - For general markdown editing
- **swagger-helper** - If `swagger.json` or `openapi.yaml` present

**Documentation structure indicators:**
- `docs/api/` - API documentation
- `docs/guides/` - User guides
- `docs/reference/` - Reference documentation
- `CONTRIBUTING.md` - Contributor guidelines

## Database Projects

**Detection patterns:**
- `migrations/` directory
- SQL files (`.sql`, `.ddl`)
- Database config files: `database.yml`, `db.json`
- ORM configs: `alembic.ini`, `knex.js`, `sequelize.config.js`
- Schema files: `schema.sql`, `schema.prisma`

**Recommended skills:**
- **postgresql-helper** - If PostgreSQL detected in configs
- **mysql-helper** - If MySQL detected in configs
- **mongodb-helper** - If MongoDB connection strings or configs present
- **prisma-helper** - If `schema.prisma` present
- **alembic-helper** - If `alembic.ini` present (Python/SQLAlchemy)
- **knex-helper** - If `knexfile.js` present (Node.js)

## Cloud & Infrastructure Projects

**Detection patterns:**
- `terraform/` directory or `.tf` files
- `ansible/` directory or `.yml` playbooks
- `cloudformation/` or CloudFormation templates
- `pulumi/` or `Pulumi.yaml`
- AWS configs: `.aws/`, `serverless.yml`
- GCP configs: `app.yaml`, `cloudbuild.yaml`

**Recommended skills:**
- **terraform-helper** - For `.tf` files
- **ansible-helper** - For Ansible playbooks
- **aws-helper** - For AWS-specific configs
- **gcp-helper** - For GCP-specific configs
- **azure-helper** - For Azure-specific configs
- **serverless-helper** - For `serverless.yml`

## Frontend Projects

**Detection patterns:**
- `public/` or `static/` directory
- `index.html` in root or `public/`
- CSS preprocessor files: `.scss`, `.sass`, `.less`
- Frontend build configs: `webpack.config.js`, `vite.config.js`, `rollup.config.js`
- Package.json with frontend frameworks

**Recommended skills:**
- **webpack-bundler** - If `webpack.config.js` present
- **vite-bundler** - If `vite.config.js` present
- **sass-helper** - If `.scss` or `.sass` files present
- **tailwind-helper** - If `tailwind.config.js` present
- **postcss-helper** - If `postcss.config.js` present

## Mobile Projects

**Detection patterns:**
- `ios/` and `android/` directories (React Native)
- `pubspec.yaml` (Flutter/Dart)
- `build.gradle` in root (Android)
- `Podfile` in root (iOS)
- `.xcodeproj` or `.xcworkspace` (iOS)

**Recommended skills:**
- **react-native-helper** - If React Native structure detected
- **flutter-helper** - If `pubspec.yaml` present
- **android-helper** - If `build.gradle` and Android structure
- **ios-helper** - If Xcode project files present

## Analysis Guidelines

### Quick Scan Approach

**DO:**
- Check only well-known indicator files in project root
- Use glob patterns for quick file existence checks
- Scan package manifests for key dependencies
- Check directory names in project root only
- Limit file reads to small config files (<100 lines)

**DON'T:**
- Read entire codebase or traverse deep directory structures
- Parse large files or analyze code complexity
- Execute build commands or run tests to detect patterns
- Make assumptions based on partial information
- Read binary files or compiled artifacts

### Prioritization Rules

1. **Root-level files first**: Files in project root are strongest indicators
2. **Multiple matches**: Recommend all matching skills, sorted by relevance
3. **Primary technology**: Identify the main language/framework first
4. **Support tools**: Add testing, CI/CD, and documentation skills second
5. **Infrastructure**: Add Docker, cloud, and database skills last

### Recommendation Template

```
Based on project analysis, I've identified the following patterns:

**Primary Stack:**
- [Language/Framework] - [Detection reason]

**Testing:**
- [Test framework] - [Detection reason]

**Infrastructure:**
- [Infrastructure tool] - [Detection reason]

**Recommended Skills:**
1. [skill-name] - [Purpose/Benefit]
2. [skill-name] - [Purpose/Benefit]
...

Would you like me to install these skills? (Yes/No/Select specific)
```

### Edge Cases

- **Monorepos**: Check for `workspaces` in package.json or `packages/` directory
- **Polyglot projects**: Multiple language indicators → recommend all relevant skills
- **Legacy projects**: Older config formats may need broader pattern matching
- **Minimal projects**: Single file projects may not match patterns → ask user for clarification

### Performance Targets

- **Analysis time**: <2 seconds for typical project
- **File reads**: <10 files for basic pattern detection
- **Glob operations**: <5 glob patterns per analysis
- **Recommendations**: 3-8 skills for typical project

## Pattern Maintenance

This reference should be updated when:
- New popular frameworks/tools emerge
- Existing tools change configuration conventions
- User feedback indicates missing or incorrect patterns
- New skills are added to the skill ecosystem

**Update frequency**: Review quarterly or when new major technologies gain >10% adoption in target communities.
