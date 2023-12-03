const express = require("express");

const app = express();
const path = require('path');
app.set("/", "html");
app.use(express.static(path.join(__dirname, "/")));
app.use(express.json());


app.use(express.urlencoded({extended: false}));
app.get('/',(request,response)=>{
    response.render("index")
})

app.listen(3000, () => {
    console.log("Listening on http://localhost:3000");
    });