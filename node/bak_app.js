"use strict";

const PORT = process.env.NODE_PORT || "40000";
const GO_URL = process.env.GO_URL || "host.docker.internal";
const GO_PORT = process.env.GO_PORT || "41000";

const express = require("express");
const axios = require("axios");

const app = express();

app.get("/node-start", (req, res) => {
  console.log("Sending call to Go!");
  axios
    .get(`http://${GO_URL}:${GO_PORT}/node-chain`)
    .then(() => axios.get(`http://${GO_URL}:${GO_PORT}/node-chain`))
    .then(result => {
      res.send(result.data);
    })
    .catch(err => {
      console.error(err);
      res.status(500).send();
    });
});

app.get("/middle-tier", (req, res) => {
  axios
    .get(`http://localhost:${PORT}/backend`)
    .then(() => axios.get(`http://localhost:${PORT}/backend`))
    .then(result => {
      res.send(result.data);
    })
    .catch(err => {
      console.error(err);
      res.status(500).send();
    });
});

app.get("/backend", (req, res) => {
  res.send("Hello from the backend");
});

app.listen(parseInt(PORT, 10), () => {
  console.log(`Listening for requests on http://localhost:${PORT}`);
});
