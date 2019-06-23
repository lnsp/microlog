pipeline {
  agent any
  environment {
      ENV_CONF = credentials('microlog-env-default')
      ENV_CONF_STAGING = credentials('microlog-env-staging')
  }
  stages {
    stage('Configure environment') {
      steps {
        sh 'echo $ENV_CONF > .env'
        sh 'echo $ENV_CONF_STAGING > .env.staging'
      }
    }
    stage('Deploy') {
      steps {
        sh 'scripts/deploy_jenkins.sh'
      }
    }
    stage('Cleanup') {
      steps {
        sh 'rm .env*'
      }
    }
  }
}
