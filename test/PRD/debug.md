# Some Debug Requirements

1. I want to visualize the traffic between Nginx and uptreams. How to? From what I know, the clients for example a browser or a curl instance teminated at Nginx then the Nginx start a new http connection to the upstream. What I want to visualized is the traffic between the Nginx and the upstream which is fired by the Nginx.
   1. Use tcpdump to intercept the traffic;
   2. The traffic is between Nginx and upstream for emample hello-server. That means you should intercept insider the docker container I think;
   3. Dump all the traffic which has both tcp and udp. Then I can open it with Wireshark to further analyze them;