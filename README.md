This is a basic prototype of Bimodal Multicast paper.
This includes 2 applications. 

The Multicaster is a simple application to generate messages with identifier.

The BMCaster is an application meant to run on each of the nodes in the cluster.
It must receive the multicast messages and gossip digests among each other.

To run the application, up.sh can be used. It will stop any existing docker containers and remove them so please edit if required.
![Bimodal Multicast Prototype - System Diagram.jpeg](https://bitbucket.org/repo/RberGR/images/2130455641-Bimodal%20Multicast%20Prototype%20-%20System%20Diagram.jpeg)

Here is a blog post which I update with current information regarding this project.
https://cdnetworks.atlassian.net/wiki/display/~Anusree/Bimodal+Multicast+Prototype
