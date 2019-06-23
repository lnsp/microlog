pipeline {
  agent any
  stages {
    stage('Deploy') {
      steps {
        sh 'microlog/scripts/deploy.sh staging'
      }
    }
  }
}