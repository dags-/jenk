function load() {
    if (window.location.pathname !== "/") {
        document.getElementById("splash").appendChild(createLoadText());
        fetch("/data" + window.location.pathname).then(r => r.json()).then(render);
    }
}

function render(data) {
    let content = document.getElementById("content");
    while (content.lastChild) {
        content.removeChild(content.lastChild);
    }

    let title = create("div", "page-title");
    title.innerText = `${data["name"]} Builds`;

    let builds = create("div", "builds-container");
    let limit = Math.min(10, data["builds"].length);
    for (let i = 0; i < limit; i++) {
        builds.appendChild(createBuild(data["builds"][i]));
    }

    content.appendChild(title);
    content.appendChild(builds);
}

function createBuild(build) {
    let date = relativeDate(new Date(build["timestamp"]));

    let title = create("div", "build-title");
    title.innerText = `Build #${pad3(build["number"])} - ${date}`;

    let commit = createCommit(build["commit"]);

    let header = create("div", "build-header");
    header.innerText = "Artifacts:";

    let artifacts = create("div", "artifacts-container");
    build["artifacts"].forEach(artifact => artifacts.appendChild(createArtifact(artifact)));

    let container = create("div", "build-container");
    container.appendChild(title);
    container.appendChild(commit);
    container.appendChild(header);
    container.appendChild(artifacts);

    return container;
}

function createCommit(commit) {
    let prefix = document.createTextNode("Commit: ");

    let link = create("a", "file-link");
    link.href = commit["url"];
    link.innerText = commit["hash"].substring(0, 8);
    link.target = "_blank";

    let container = create("div", "build-header");
    container.appendChild(prefix);
    container.appendChild(link);

    return container;
}

function createArtifact(artifact) {
    let link = create("a", "file-link");
    link.href = artifact["relativePath"];
    link.innerText = artifact["fileName"];

    let container = create("div", "file-container");
    container.appendChild(link);

    return container;
}

function createLoadText() {
    let div = create("div");
    div.innerText = `Loading...`;
    return div;
}

function create(type, style) {
    let el = document.createElement(type);
    el.classList.add(style);
    return el;
}

function relativeDate(date) {
    let delta = Math.round((+new Date - date) / 1000);
    let minute = 60, hour = minute * 60, day = hour * 24;
    if (delta < 30) {
        return "just now";
    } else if (delta < minute) {
        return delta + " seconds ago";
    } else if (delta < 2 * minute) {
        return "a minute ago"
    } else if (delta < hour) {
        return Math.floor(delta / minute) + " minutes ago";
    } else if (Math.floor(delta / hour) === 1) {
        return "an hour ago"
    } else if (delta < day) {
        return Math.floor(delta / hour) + " hours ago";
    } else if (delta < day * 2) {
        return "yesterday";
    }
    return `on ${date.getFullYear()}/${pad2(date.getMonth()+1)}/${pad2(date.getDate())}`;
}

function pad2(number) {
    if (number < 10) {
        return "0" + number;
    }
    return number;
}

function pad3(number) {
    if (number < 10) {
        return "00" + number;
    }
    if (number < 100) {
        return "0" + number;
    }
    return number;
}