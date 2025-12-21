pipeline {
  agent any

  environment {
    IMAGE_NAME = 'reyhandhani11/temuin'
    REGISTRY_CREDENTIALS = 'dockerhub-credentials'
  }

  stages {

    stage('Checkout') {
      steps {
        checkout scm
      }
    }

    stage('Build Docker Image') {
      steps {
        script {
          docker.build("${IMAGE_NAME}:${BUILD_NUMBER}")
        }
      }
    }

    stage('Push Docker Image') {
      steps {
        script {
          docker.withRegistry('', REGISTRY_CREDENTIALS) {
            docker.image("${IMAGE_NAME}:${BUILD_NUMBER}").push()
            docker.image("${IMAGE_NAME}:${BUILD_NUMBER}").tag('latest')
            docker.image("${IMAGE_NAME}:latest").push()
          }
        }
      }
    }
  }
}
