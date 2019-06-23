pipeline {
  agent any
  stages {
    stage('Deploy') {
      steps {
        sh 'scripts/deploy.sh staging'
      }
    }
  }
}