pipeline {
    agent { label 'java-docker-slave' }

    environment {
        DOCKER_IMAGE_NAME = "bookapp"
    }

    stages {
        stage("Checkout") {
            steps {
              git url: 'https://github.com/jenoOvchi/bookapp' 
              }
            }

        stage("Docker Build and Test") {
            steps {
                script {
                    docker.withServer('tcp://192.168.10.3:4243') {
                        app = docker.build(DOCKER_IMAGE_NAME)
                        docker.withRegistry('http://192.168.10.3:5000') {
                            app.push("${env.BUILD_NUMBER}")
                            app.push("latest")
                        }
                    }
                }
            }
        }

        stage("Deploy Image") {
            steps {
                milestone(1)
                sleep 15
                echo "Some deployment steps"
            }
        }
    }
}
