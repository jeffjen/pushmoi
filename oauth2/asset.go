package oauth2

const (
	OAUTH2_WORKFLOW_HTML = `
<html>
<head>
<title>Pushmoi</title>
<script src="//ajax.googleapis.com/ajax/libs/angularjs/1.5.7/angular.min.js"></script>
<script src="//ajax.googleapis.com/ajax/libs/angularjs/1.5.7/angular-resource.js"></script>
</head>
<body ng-app="pushmoi">
  <pushmoi-auth></pushmoi-auth>
</body>
<script>
angular.
  module("pushmoi", ["ngResource"]).
  factory("Token", ["$resource",
    function ($resource) {
      return $resource("/pushmoi/authroized");
    }
  ]).
  component("pushmoiAuth", {
    template:
      '<div ng-style="$ctrl.containterStyle">' +
      '  <div ng-if="$ctrl.error">{{ $ctrl.error }}</div>' +
      '  <div ng-if="!$ctrl.error">' +
      '    <div>Obtained PushBullet access_token:</div>' +
      '    <div ng-style="$ctrl.tokenStyle">{{ $ctrl.access_token }}</div>' +
      '  </div>' +
      '</div>',
    controller: ["Token", "$location",
      function (Token, $location) {
        this.containterStyle = {
          "font-family": "monospace"
        };
        this.tokenStyle = {
          "background-color": "#ccc",
		  display: "inline-block",
		  "margin-top": "5px",
          padding: "5px 15px"
        };
        var [key, value] = $location.path().split("=");
        if (key === "/error") {
          this.error = error;
        } else if (key === "/access_token") {
          this.access_token = value;
          Token.save({ access_token: this.access_token }, () => {
            // NOTE: report token store success
          }, (error) => {
            this.error = error;
          });
        }
      }
    ]
  });
</script>
</html>
`
)
