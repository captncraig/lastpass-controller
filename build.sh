export KO_DOCKER_REPO=captncraig

ko publish --platform=all -B -t `git rev-parse HEAD` .
ko publish --platform=all -B .