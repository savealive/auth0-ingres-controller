pipeline {
  agent {
    label "jenkins-go"
  }
  environment {
    ORG = 'creditplace'
    APP_NAME = 'auth0-ingress-controller'
    CHARTMUSEUM_CREDS = credentials('jenkins-x-chartmuseum')
    DOCKER_REGISTRY_ORG = 'creditplace'
    gitShortCommit = sh(
            script: "printf \$(git rev-parse --short ${GIT_COMMIT})",
            returnStdout: true, label: "get short commit"
    ).trim()
  }
  stages {
    stage('CI Build and push snapshot') {
      when {
        branch 'PR-*'
      }
      environment {
        PREVIEW_VERSION = "0.0.0-SNAPSHOT-$BRANCH_NAME-$BUILD_NUMBER"
        PREVIEW_NAMESPACE = "$APP_NAME-$BRANCH_NAME".toLowerCase()
        HELM_RELEASE = "$PREVIEW_NAMESPACE".toLowerCase()
      }
      steps {
        container('go') {
          checkout scm
          sh "make linux"
          sh "export VERSION=$PREVIEW_VERSION && skaffold build -f skaffold.yaml"
          sh "jx step post build --image $DOCKER_REGISTRY/$ORG/$APP_NAME:$PREVIEW_VERSION"
          // dir('charts/preview') {
          //   sh "make preview"
          //   //sh "jx preview --app $APP_NAME --dir ../.."
          // }
        }
      }
    }
    stage('Build develop') {
      when {
        branch 'dev'
      }
      environment {
        VERSION = "$BRANCH_NAME-$gitShortCommit-b$BUILD_NUMBER"
      }
      steps {
        container('go') {
            checkout scm

            // ensure we're not on a detached head
            sh "git checkout dev"
            sh "git config --global credential.helper store"
            sh "jx step git credentials"
            // so we can retrieve the version in later steps
            sh "make build"
            sh "skaffold build -f skaffold.yaml"
        }
      }
    }
    stage('Build Release') {
      when {
        branch 'master'
      }
      steps {
        container('go') {
            checkout scm

            // ensure we're not on a detached head
            sh "git checkout master"
            sh "git config --global credential.helper store"
            sh "jx step git credentials"

            // so we can retrieve the version in later steps
            sh "jx step next-version --use-git-tag-only -t"
            sh "make build"
            sh "export VERSION=`cat VERSION` && skaffold build -f skaffold.yaml"
            sh "jx step post build --image $DOCKER_REGISTRY/$ORG/$APP_NAME:\$(cat VERSION)"
        }
      }
    }
    stage('Promote to Environments') {
      when {
        branch 'master'
      }
      steps {
        container('go') {
          dir('charts/auth0-ingress-controller') {
            sh "jx step changelog --version v\$(cat ../../VERSION)"

            // release the helm chart
            sh "jx step helm release"

            // promote through all 'Auto' promotion Environments
            //sh "jx promote -b --all-auto --timeout 1h --version \$(cat ../../VERSION)"
          }
        }
      }
    }
  }
}
