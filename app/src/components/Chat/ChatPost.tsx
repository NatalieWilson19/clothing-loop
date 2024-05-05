import { IonActionSheet, IonItem, useIonAlert } from "@ionic/react";
import { Post } from "@mattermost/types/posts";
import { t } from "i18next";
import { useState } from "react";
import { useLongPress } from "use-long-press";

interface Props {
  post: Post;
  isMe: boolean;
}

export default function ChatPost({ post, isMe }: Props) {
  const [isActionSheetOpen, setIsActionSheetOpen] = useState(false);
  const [presentAlert] = useIonAlert();

  const longPressMessage = useLongPress(
    (e) => {
      setIsActionSheetOpen(true);
    },
    {
      onCancel: (e) => {
        setIsActionSheetOpen(false);
      },
    },
  );
  function onDeleteMessageSubmit() {
    console.log("delete message: ", post.message);
  }
  function onReportMessageSubmit() {
    console.log("report message", post.message);
  }
  function handleOptionSelect(value: string) {
    if (value == "delete") {
      const handler = () => {
        onDeleteMessageSubmit();
      };
      presentAlert({
        header: "Delete message?",
        buttons: [
          {
            text: t("cancel"),
          },
          {
            role: "destructive",
            text: t("delete"),
            handler,
          },
        ],
      });
    } else if (value == "report") {
      const handler = () => {
        onReportMessageSubmit();
      };
      presentAlert({
        header: "Report message?",
        buttons: [
          {
            text: t("cancel"),
          },
          {
            role: "destructive",
            text: t("report"),
            handler,
          },
        ],
        inputs: [
          {
            placeholder: "Report Description (Optional)",
            type: "textarea",
          },
        ],
      });
    }
  }

  return (
    <div>
      <IonItem
        lines="none"
        color="light"
        className={`tw-shrink-0 tw-rounded-tl-2xl tw-rounded-tr-2xl ${
          post.is_following ? "" : "tw-mb-2"
        } ${
          isMe
            ? "tw-rounded-bl-2xl tw-float-right tw-ml-8 tw-mr-4"
            : "tw-rounded-br-2xl tw-mr-8 tw-ml-4 tw-float-left"
        }`}
        {...longPressMessage()}
      >
        <div className="tw-py-2">
          <div className="tw-font-bold">{post.props.username}</div>
          <div>{post.message}</div>
        </div>
      </IonItem>
      <IonActionSheet
        header={t("actions")}
        isOpen={isActionSheetOpen}
        onDidDismiss={() => setIsActionSheetOpen(false)}
        buttons={[
          {
            text: isMe ? "Delete" : "Report",
            handler: () => handleOptionSelect(isMe ? "delete" : "report"),
          },
          {
            text: "Cancel",
            role: "cancel",
          },
        ]}
      ></IonActionSheet>
    </div>
  );
}
