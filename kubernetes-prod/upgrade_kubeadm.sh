K8S_VERSION_MINOR=1.30

# If the directory `/etc/apt/keyrings` does not exist, it should be created before the curl command, read the note below.
# sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v${K8S_VERSION_MINOR}/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v${K8S_VERSION_MINOR}/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list

# Upgrade kubeadm
sudo apt-mark unhold kubeadm && \
sudo apt-get update && sudo apt-get install -y kubeadm='1.30.2-1.1' && \
sudo apt-mark hold kubeadm

# sudo kubeadm upgrade node

# # Drain the node 
# kubectl drain $node_name --ignore-daemonsets

# # Upgrade the kubelet and kubectl
# sudo apt-mark unhold kubelet kubectl && \
# sudo apt-get update && sudo apt-get install -y kubelet='1.30.2-1.1' kubectl='1.30.2-1.1' && \
# sudo apt-mark hold kubelet kubectl

# # Restart the kubelet
# sudo systemctl daemon-reload
# sudo systemctl restart kubelet

# # Uncordon the node
# kubectl uncordon $node_name