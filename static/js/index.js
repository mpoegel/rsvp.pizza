// let guestLevels = document.getElementsByClassName("guest-level");
// for (let i = 0; i < guestLevels.length; i++) {
//     guestLevels[i].addEventListener("mouseDown", function (e) {
//         console.log(e);
//     });
// }

function expandGuests(event) {
    // console.log(event);
    if (event.className === "guest-level") {
        // expand
        event.className = "guest-level-expanded";
        for (let i = 0; i < event.children.length; i++) {
            event.children[i].className = "guest-expanded";
            event.children[i].children[1].className = "guest-name-expanded";
        }
    } else {
        // collapse
        event.className = "guest-level";
        for (let i = 0; i < event.children.length; i++) {
            event.children[i].className = "guest";
            event.children[i].children[1].className = "guest-name";
        }
    }
}
