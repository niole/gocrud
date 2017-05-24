var express = require('express');
var bodyParser = require('body-parser');
var cookieParser = require('cookie-parser');
var path = require('path');
var app = express();
var server = require('http').Server(app);
var mysql = require('mysql')

app.use(cookieParser());
app.use(bodyParser.urlencoded({
    extended: true
}));

var connection = mysql.createConnection({
  host     : '127.0.0.1',
  user     : 'root',
  password : 'root',
  database : 'mysql',
  port: '3307',
});

connection.connect()

app.get('/index', function (req, res) {
  res.sendFile(path.join(__dirname, '../public/index.html'));
});

app.post('/todo/create', function (req, res) {
  var b = JSON.parse(Object.keys(req.body)[0]);

  var fields = Object.keys(b);
  var values = fields.map(function(field) {
    if (typeof b[field] === "string") {
      return "'"+b[field]+"'";
    }
    return b[field];
  });

  var query = 'INSERT INTO todo ('+ fields.join(",") + ") VALUES ("+ values.join(",") +")";
  connection.query(query, function (err, rows, fields) {
    if (err) throw err

    res.send(rows)
  });

});

app.post('/todo/read', function (req, res) {
  var b = JSON.parse(Object.keys(req.body)[0]);
  var whereClause = b.where;
  var whereStatement = "("+Object.keys(whereClause).map(function(col) {
    return col+"="+whereClause[col];
  }).join(",")+")";

  connection.query('SELECT * FROM todo WHERE '+whereStatement, function (err, rows, fields) {
    if (err) throw err
    res.send(JSON.stringify(rows));
  });

});

var app_port = process.env.PORT || 8080;
server.listen(app_port, function() {
  console.log('listening to port ' + app_port);
});

