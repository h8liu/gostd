package main

const htmlHeader = `<!doctype html>
<html>

<head>
	<meta charset="utf-8" />
	<title>Go standard library</title>
	<link rel="stylesheet" type="text/css" href="/gostd.css">
</head>

<body>

`

const htmlFooter = `
<script>
  (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
  (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
  m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
  })(window,document,'script','//www.google-analytics.com/analytics.js','ga');

  ga('create', 'UA-19116218-3', 'auto');
  ga('send', 'pageview');
</script>
</body>
</html>
`
