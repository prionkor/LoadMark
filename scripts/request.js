import http from "k6/http";

const method = __ENV.METHOD;
const url = __ENV.URL;
const body = __ENV.BODY;
const headers = JSON.parse(__ENV.HEADERS);
const executor = __ENV.EXECUTOR;
const executorSettings = JSON.parse(__ENV.EXECUTOR_SETTINGS);

export const options = {
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
  scenarios: {
    load_test: buildLoadTest(executor, executorSettings),
  },
};

export default function () {
  http.request(method, url, body, {
    headers,
  });
}

function buildLoadTest(executor, settings) {
  const loadTest = {
    executor,
  };

  if (settings.start_rps !== undefined) {
    loadTest.startRate = settings.start_rps;
  }

  if (settings.time_unit !== undefined) {
    loadTest.timeUnit = settings.time_unit;
  }

  if (settings.pre_allocated_vus !== undefined) {
    loadTest.preAllocatedVUs = settings.pre_allocated_vus;
  }

  if (settings.max_vus !== undefined) {
    loadTest.maxVUs = settings.max_vus;
  }

  if (
    settings.stages !== undefined &&
    Array.isArray(settings.stages) &&
    settings.stages.length > 0
  ) {
    loadTest.stages = settings.stages.map((stage) => ({
      target: stage.rps,
      duration: stage.duration,
    }));
  }

  return loadTest;
}
