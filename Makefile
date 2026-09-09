.PHONY: tidy daemon coordinator bridge llama all

# Compilação pendente: rodar `make all` quando for a hora.
all: tidy daemon coordinator bridge

tidy:
	cd services/daemon && go mod tidy
	cd services/coordinator && go mod tidy
	cd spikes/fase0/bridge && go mod tidy

daemon:
	cd services/daemon && go build -o bin/p2pai-daemon ./cmd/p2pai-daemon

coordinator:
	cd services/coordinator && go build -o bin/p2pai-coordinator ./cmd/p2pai-coordinator

bridge:
	cd spikes/fase0/bridge && go build -o bridge .

# llama.cpp com RPC + Metal (fonte clonado em spikes/fase0/llama.cpp, ignorado pelo git)
llama:
	cd spikes/fase0/llama.cpp && cmake -B build -DGGML_RPC=ON -DGGML_METAL=ON -DLLAMA_CURL=OFF -DCMAKE_BUILD_TYPE=Release \
		&& cmake --build build --config Release -j 8 --target ggml-rpc-server llama-completion llama-server llama-bench
