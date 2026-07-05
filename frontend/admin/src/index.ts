import { defineSubsystem } from "@bit-labs.cn/owl-ui";
import routes from "./routes";

export default defineSubsystem({
  name: "metrics",
  routes,
  viewModulesPathPrefix: "/metrics",
  viewModules: import.meta.glob("./views/**/*.{vue,tsx}"),
  menuContributions: []
});
