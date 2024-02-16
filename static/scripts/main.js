document.addEventListener("DOMContentLoaded", Init)

const expressionsList = document.querySelector(".expressions-list")
const workersList = document.querySelector(".workers-list")
let expressions = []
let workers = []
let delays = {}
let windows = document.querySelectorAll(".window")
const modal = document.querySelector(".modal")

function Init(){
    getExpressions().then(data => {expressions = data
        for (const i of expressions){
            expressionsList.prepend(CreateExpression(i))
        }})
    getWorkers().then(data => {workers = data
        for (const i of workers){
            workersList.append(CreateWorker(i))
        }})
    getDelays().then(data => {delays = data
        for (let i in delays){
            document.getElementById(i).value = delays[i]
        }
    })
    for (const i of document.querySelectorAll(".menu-btn")){
        i.addEventListener("click", ChangeWindow)
        i.wind = document.querySelector("."+i.dataset.show)
    }
    document.querySelector(".expression-btn").addEventListener("click", sendExpression)
    document.querySelector(".operations-btn").addEventListener("click", saveDelays)
    UpdateData()
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

async function UpdateData(){
    while (true){
        await sleep(6000)
        getExpressions().then(data => {expressions = data
            const updated = document.createElement("div")
            for (const i of expressions){
                updated.prepend(CreateExpression(i))
            }expressionsList.innerHTML = updated.innerHTML})
        getWorkers().then(data => {workers = data
            const updated = document.createElement("div")
            for (const i of workers){
                updated.append(CreateWorker(i))
            } workersList.innerHTML=updated.innerHTML})
    }
}

function getExpressionInfo(id){
    modal.style.display = "block"
    getTask(id).then(data => showExpressionInfo(data)).catch(error => showNotification(error))
}

function showExpressionInfo(task){
    const content = document.querySelector(".modal-content")
    const list = document.querySelector(".subtasks-list")
    list.innerHTML = ""
    if (task["subtasks"]){
        for (const i of task["subtasks"]){
            const li = document.createElement("li")
            li.innerHTML = `<pre>${i["value"]} &rarr; ${i["result"]}            ${i["time"]}</pre>`
            list.append(li)
        }
    }
    content.querySelector(".expression-id").innerText = "ID: " + task["id"]
    content.querySelector(".expression-status").innerText = "Статус: " + task["status"]
    content.querySelector(".expression-lastPing").innerText = "Последнее обновление: " + task["lastPing"]
    content.querySelector(".expression-created").innerText = "Создан: " + task["created"]
    content.querySelector(".expression-result").innerText = "Результат: " + task["result"]
    content.querySelector(".expression-info-value").innerText = task["expression"]
}

function getTask(id){
    const data = fetch(window.location.origin + "/getTask", {
        method: "POST",
        body: JSON.stringify({"id": +id}),
        headers: {
            "Content-Type": "application/json"
        }
    }).then(response => {if (!response.ok) {
        return response.text().then(text => Promise.reject(text));
    }
    return response.json();})
    return data
}

window.onclick = function(event) {
    if (event.target == modal) {
      modal.style.display = "none";
    }
    if (event.target.classList.contains("delete-btn")){
        DeleteWorker(+event.target.dataset.id)
    }
    if (event.target.classList.contains("moreinfo")){
        getExpressionInfo(+event.target.dataset.id)
    }
  }

function saveDelays(){
    const operations = {}
    for (const i of document.querySelectorAll(".operation")){
        operations[i.id] = +i.value
    }
    fetch(window.location.origin + "/updateDelays",{
        body: JSON.stringify({"delays": operations}),
        method: "POST",
        headers:{
            "Content-Type": "application/json"
        }
    }).then(response => {
        if (!response.ok) {
            return response.text().then(text => Promise.reject(text));
        }
        return response.json();
    })
    .then(data => {
        showNotification("Обновлено");
        delays = data;
    })
    .catch(error => {
        showNotification(error);
    });
}

async function getDelays(){
    const url = window.location.origin + "/getDelays"
    const response = await fetch(url)
    const data = await response.json()
    return data
}

function sendExpression(){
    const expression = document.querySelector(".expression-value").value
    if (!expression.trim()){
        showNotification("Выражение пустое")
        return
    }
    fetch(window.location.origin + "/addTask",{
        body: JSON.stringify({"task": expression}),
        method: "POST",
        headers:{
            "Content-Type": "application/json"
        }
    }).then(response => {
        if (!response.ok) {
            return response.text().then(text => Promise.reject(text));
        }
        return response.json();
    })
    .then(data => {
        showNotification("Выражение принято");
        expressionsList.prepend(CreateExpression(data));
    })
    .catch(error => {
        showNotification(error);
    });
}

function showNotification(text) {
    const notification = document.getElementById("notification");
    notification.innerText = text
    notification.className = "notification show";
    setTimeout(function() {
      notification.className = notification.className.replace("show", "");
    }, 3000);
  }

async function getExpressions(){
    const url = window.location.origin + "/getAllExpressions"
    const response = await fetch(url)
    const data = await response.json()
    return data
}

async function getWorkers(){
    const url = window.location.origin + "/getWorkersInfo"
    const response = await fetch(url)
    const data = await response.json()
    return data
}

function CreateWorker(worker) {
    const result = document.createElement("li")
    result.classList.add("workers-list-item")
    const div = document.createElement("div")
    div.dataset.id = worker["id"]
    div.innerHTML = `<p class="worker-name">Worker_${worker["id"]}</p>`
    const workerinfo = document.createElement("div")
    workerinfo.classList.add("worker-info")
    workerStatus = document.createElement("p")
    workerStatus.innerHTML = `Статус: ${worker["status"]}`
    const btn = document.createElement("button")
    btn.innerText = "Удалить"
    btn.dataset.id = worker["id"]
    btn.classList.add("delete-btn")
    workerinfo.append(workerStatus)
    if (worker["expression"] != ""){
        workerExpression = document.createElement("p")
        workerExpression.innerText = worker["expression"]
        workerExpression.innerHTML = `Подзадача: ${worker["expression"]}`
        workerExpressionId = document.createElement("p")
        workerExpressionId.innerHTML = `ID задачи: ${worker["expressionId"]}`
        workerinfo.append(workerExpression)
        workerinfo.append(workerExpressionId)
    }
    workerinfo.append(btn)
    div.append(workerinfo)
    result.append(div)
    return result
}

function DeleteWorker(id){
    fetch(window.location.origin + "/removeWorker",{
        body: JSON.stringify({"id": id}),
        method: "POST",
        headers:{
            "Content-Type": "application/json"
        }
    }).then(response => {
        if (!response.ok) {
            return response.text().then(text => Promise.reject(text));
        }
        return response.text();
    })
    .then(data => {
        showNotification("Удален");
    })
    .catch(error => {
        showNotification(error);
    });
}

function CreateExpression(expression) {
    const result = document.createElement("li") 
    const div = document.createElement("div")
    value = document.createElement("p")
    value.classList.add("expression-value")
    value.innerText = expression["expression"]
    if (expression["status"] == "completed"){
        value.innerText += "=" + expression["result"]
    }
    div.dataset.id = expression["id"]
    div.append(value)
    const info = document.createElement("div")
    info.insertAdjacentHTML("beforeend", `<p class="id">ID: ${expression["id"]}</p>`)
    info.insertAdjacentHTML("beforeend", `<p class="lastPing">Последнее обновление: ${expression["lastPing"]}</p>`)
    info.insertAdjacentHTML("beforeend", `<p class="moreinfo" data-id="${expression["id"]}">Больше информации</p>`)
    info.classList.add("info")
    div.append(info)
    div.classList.add(expression["status"])
    div.classList.add("item")
    result.append(div)
    return result
}

function ChangeWindow(){
    for (const i of windows){
        if (!i.classList.contains("hide")){
            i.classList.add("hide")
        }
    }
    this.wind.classList.remove("hide")
}
