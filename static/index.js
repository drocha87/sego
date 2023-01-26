console.log("hello");

const input = document.querySelector("#query");

// TODO: implement debouncer

async function search() {
  try {
    const value = input.value;
    if (value !== "") {
      const url = `http://localhost:6969/search?query=${value}`;

      const response = await fetch(encodeURI(url), {
        method: "GET",
      });
      if (response.status === 200) {
        const data = await response.json();
        const { length, results } = data;

        const resultsContainer = document.querySelector("#results");
        while (resultsContainer.firstChild) {
          resultsContainer.removeChild(resultsContainer.lastChild);
        }

        if (length && length > 0) {
          for (let result of results
            .sort((a, b) => b.freq - a.freq)
            .slice(0, 10)) {
            const freqSpan = document.createElement("span");
            freqSpan.classList.add("result-freq");
            freqSpan.innerText = result.freq.toFixed(4);

            const pathSpan = document.createElement("span");
            pathSpan.classList.add("result-path");
            pathSpan.innerText = result.path;

            const el = document.createElement("div");
            el.classList.add("result-item");
            el.appendChild(pathSpan);
            el.appendChild(freqSpan);

            resultsContainer.appendChild(el);
          }
        }

        input.value = "";
      }
    }
  } catch (error) {
    console.log(error);
  }
}
