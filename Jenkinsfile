pipeline {
  agent any

  environment {
    PNPM_VERSION = '11.23.0'
    SQLC_VERSION = 'v1.29.0'
    IMAGE_TAG = "${env.GIT_COMMIT}"
  }

  stages {
    stage('Backend quality') {
      steps {
        dir('backend') {
          sh 'test -z "$(gofmt -l .)"'
          sh 'go vet ./...'
          sh 'go run honnef.co/go/tools/cmd/staticcheck@v0.6.1 ./...'
          sh 'go test ./...'
          sh 'go test -race ./...'
        }
      }
    }

    stage('Generation verification') {
      steps {
        dir('backend') {
          sh 'go run github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION} generate'
          sh 'git diff --exit-code -- internal/db'
        }
      }
    }

    stage('PostgreSQL integration') {
      steps {
        dir('backend') {
          sh 'go test -tags=integration -v ./tests/integration'
        }
      }
    }

    stage('Frontend quality') {
      steps {
        dir('frontend') {
          sh 'npx --yes pnpm@${PNPM_VERSION} install --frozen-lockfile'
          sh 'npx --yes pnpm@${PNPM_VERSION} lint'
          sh 'npx --yes pnpm@${PNPM_VERSION} typecheck'
          sh 'npx --yes pnpm@${PNPM_VERSION} test'
          sh 'npx --yes pnpm@${PNPM_VERSION} build'
        }
      }
    }

    stage('Immutable images') {
      steps {
        sh 'test -n "${IMAGE_TAG}"'
        sh 'docker build --build-arg APP_VERSION=${IMAGE_TAG} -t aurora-shop-backend:${IMAGE_TAG} backend'
        sh 'docker build --build-arg APP_VERSION=${IMAGE_TAG} -t aurora-shop-frontend:${IMAGE_TAG} frontend'
      }
    }
  }

  post {
    always {
      junit allowEmptyResults: true, testResults: '**/test-results/*.xml'
    }
  }
}
