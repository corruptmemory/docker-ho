#!/usr/bin/env bash

function check_available() {
  which $1 >/dev/null
  if [ $? -ne 0 ]; then
    echo "**** ERROR needed program missing: $1"
    exit 1
  fi
}

check_available 'which'
check_available 'realpath'
check_available 'dirname'
check_available 'go'
check_available 'docker'

script_dir=$(dirname "$(realpath -e "$0")")
cwd="$(echo "$(pwd)")"
function cleanup() {
  cd "$cwd"
}
# Make sure that we get the user back to where they started
trap cleanup EXIT

# This is necessary because we reference things relative to the script directory
cd "$script_dir"

function usage() {
  echo "Usage: build.sh [-h|--help] [-c|--clean] [-C|--clean-all]"
  echo "                [-b|--build] [-r|--run]"
  echo "                [-D|--docker] [-R|--docker-run]"
  echo
  echo '    Build docker-ho.'
  echo
  echo "Arguments:"
  echo "  -h|--help               This help text"
  echo '  -c|--clean              Clean generated artifacts.'
  echo "  -C|--clean-all          Clean all the artifacts and the Go module cache."
  echo "  -b|--build              Build 'docker-ho' using local tooling"
  echo "  -r|--run                Build and run 'docker-ho' using local tooling"
  echo "  -D|--docker             Build a docker image of 'docker-ho'"
  echo "  -R|--docker-run         Run the docker image of 'docker-ho'"
}

clean=0
clean_all=0
build=0
run=0
create_docker=0
run_docker=0

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
  -h | --help)
    usage
    exit 0
    ;;
  -c | --clean)
    clean=true
    shift
    ;;
  -C | --clean-all)
    clean_all=true
    shift
    ;;
  -b | --build)
    build=true
    shift
    ;;
  -r | --run)
    run=true
    shift
    ;;
  -D | --docker)
    create_docker=true
    shift
    ;;
  -R | --docker-run)
    run_docker=true
    shift
    ;;
  *)
    echo "ERROR: unknown argument $1"
    echo
    usage
    exit 1
    ;;
  esac
done

docker_out="docker_out"

if [ "$clean_all" = true ]; then
  echo "Deep cleaning..."
  clean=true
  go clean --modcache
fi

if [ "$clean" = true ]; then
  echo "Regular cleaning..."
	rm -fr "./${docker_out}"
	rm -f docker-ho
	go clean .
	docker rmi -f docker-ho:latest
fi

if [ "$build" = true ] || [ "$run" = true ]; then
  echo "Building..."
	CGO_ENABLED=0 go build -v -o "docker-ho"
fi

if [ "$run" = true ]; then
  ./docker-ho
fi

if [ "$create_docker" = true ] || [ "$run_docker" = true ]; then
  echo "Creating docker image..."
  rm -fr "./${docker_out}"
  mkdir "./${docker_out}"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o "${docker_out}/docker-ho"
  docker build --tag docker-ho:latest .
fi

if [ "$run_docker" = true ]; then
  echo "Running docker image..."
  docker run --rm -p 8080:8080 docker-ho
fi
