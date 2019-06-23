pipeline {
  agent any
  environment {
      ENV_CONF = credentials('microlog-env-default')
      ENV_CONF_STAGING = credentials('microlog-env-staging')
  }
  stages {
    stage('Configure environment') {
      steps {
        sh 'cp $ENV_CONF .env'
        sh 'cp $ENV_CONF_STAGING .env.staging'
      }
    }
    stage('Deploy') {
      when { branch: 'master' }
      steps {
        sh 'scripts/deploy_jenkins.sh'
      }
    }
    post {
      always {
        deleteDir()
      }
    }
  }
}
