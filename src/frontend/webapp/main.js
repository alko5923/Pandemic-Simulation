/* When the user clicks on the button, 
toggle between hiding and showing the dropdown content */

let setting = "";

function toggleSettings(menu) 
{
    if(menu == setting)
    {
        $("#" + setting).toggleClass("show");
        setting = "";
        return;
    }

    if(setting != "")
        $("#" + setting).toggleClass("show");

    $("#" + menu).toggleClass("show");
    setting = menu;
}