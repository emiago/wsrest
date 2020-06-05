# WSREST: HTTP framework over websocket

Similar framework I used in production.
Goal of this framework is to work with websocket streams and 
request response model with focus o to get smaller network traffic, and avoid creation new connections. Plus it must allow handling 
standard HTTP without having to need changing code. 

This kind of framework is good where you have services that 
constantly are talking to each other, but you also have web GUI that can configure or monitor your service using either HTTP or Websocket.

Now as websocket is duplex communication, there is implementation of simple protocol that mimics http request response. In this case req and resp are smaller packets than standard http. 

Here example req resp as JSON packets:
```
Request:
{ 
    uid: "req1",
    m : "GET",  //Method : GET, POST, PUT, DELETE
    r : "/myresource"
}

Response:
{
    uid: "req1",
    m : "GET", 
    r : "/myresource"
    c : 200,  //http status code
    d : {...some json data ...}
}
```

Going a bit further more today services are event driven, so this is also part of framework. There is simple implementation of **pubsub** but follows this concepts: 
- events must be handled serial
- subscribe and unsubscribe must be possible as async operation
- dynamically subscribe and unsubscribe must be possible, where unsubscribe will not discard any prior published events


