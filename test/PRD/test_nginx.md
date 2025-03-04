# Test The Nginx Build
Now the Nginx build is done. Next stage I am going to test the functionality of the Nginx build. 

1. So I wanna run a helloworld babby service and let the nginx reverse proxy the service.
2. Route the flow starting with `/api` to the helloworld server;
3. Test file streaming: 
   1. with route `/api/file/$uuid`; 
   2. Download with `GET` method;
   3. Upload with same point;
4. Test WS streaming;


## Reqirements

1. Use golang to implement;
2. Use Docker Compose for simplifing the whole process and consistence among many enviroments;
3. Use new docker compose specification. Like using `docker compose` instead of `docker-compose`;
4. Use the Nginx build from the root directory which has a `Dockerfile`;
5. Update the update-to-date info to README.md file;