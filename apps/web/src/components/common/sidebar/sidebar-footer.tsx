import BottomActions from "./sections/bottom-actions";
import UserInfo from "./sections/user-info";

function SidebarFooter() {
  return (
    <div className="p-4 border-t border-zinc-200 dark:border-zinc-800">
      <UserInfo />
      <BottomActions />
    </div>
  );
}

export default SidebarFooter;
