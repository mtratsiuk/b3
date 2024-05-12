(function () {
  var DARK_THEME_CLASSNAME = "dark";

  var themeToggle = document.getElementById("theme-toggle");
  var state = themeToggle.textContent

  themeToggle.addEventListener("click", function () {
    document.documentElement.classList.toggle(DARK_THEME_CLASSNAME);

    state = state === "jedi" ? "sith" : "jedi"

    themeToggle.textContent = state
  });
})();
