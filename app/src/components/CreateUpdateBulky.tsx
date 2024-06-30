import {
  IonButton,
  IonButtons,
  IonCard,
  IonContent,
  IonHeader,
  IonIcon,
  IonImg,
  IonInput,
  IonItem,
  IonLabel,
  IonList,
  IonModal,
  IonTextarea,
  IonTitle,
  IonToolbar,
  isPlatform,
  useIonToast,
} from "@ionic/react";
import type { IonModalCustomEvent } from "@ionic/core";
import {
  checkmarkOutline,
  cloudUploadOutline,
  hourglassOutline,
  imageOutline,
} from "ionicons/icons";
import { ChangeEvent, RefObject, useContext, useRef, useState } from "react";
import { bulkyItemPut } from "../api/bulky";
import { BulkyItem } from "../api/types";
import { StoreContext } from "../stores/Store";
import { OverlayEventDetail } from "@ionic/react/dist/types/components/react-component-lib/interfaces";
import toastError from "../../toastError";
import { useTranslation } from "react-i18next";
import { Camera, CameraResultType } from "@capacitor/camera";
import { uploadImage } from "../api/imgbb";

enum State {
  idle,
  error,
  success,
  loading,
}

export default function CreateUpdateBulky({
  type,
  bulky,
  didDismiss,
  modal,
  onUploadBulkyImage,
  onSendBulkyItem,
  onUpdateBulkyItem,
  postID,
  title,
  message,
  fileID,
}: {
  type: "create" | "update";
  bulky: BulkyItem | null;
  modal: RefObject<HTMLIonModalElement>;
  didDismiss?: (
    e: IonModalCustomEvent<OverlayEventDetail<BulkyItem | null>>,
  ) => void;
  onUploadBulkyImage :(
    image: File
  ) => Promise<string| undefined>;
  onSendBulkyItem: (
    message: string,
    fileID: string,
    callback: Function,
  ) => Promise<void>;
  onUpdateBulkyItem: (
    postId: string,
    message: string,
    callback: Function,
  ) => Promise<void>;
  postID?: string;
  title?: string;
  message?: string;
  fileID?: string;
}) {
  const { t } = useTranslation();
  const { chain, authUser } = useContext(StoreContext);
  const [bulkyTitle, setBulkyTitle] = useState(title);
  const [bulkyMessage, setBulkyMessage] = useState(message);
  const [bulkyImageURL, setBulkyImageURL] = useState("");
  const [error, setError] = useState("");
  const [isCapacitor] = useState(() => isPlatform("capacitor"));
  const [loadingUpload, setLoadingUpload] = useState(State.loading);
  const [present] = useIonToast();
  const refScrollRoot = useRef<HTMLDivElement>(null);

  const [imageFile, setImageFile] = useState<File>();

  function modalInit() {
    console.log(bulkyTitle, bulkyMessage);
    const url = fileID ? `http://localhost:8065/api/v4/files/${fileID}/preview` : ""

    setBulkyTitle(title || "");
    setBulkyMessage(message || "");
    setBulkyImageURL(url);
    setLoadingUpload(State.idle);
  }

  function cancel() {
    modal.current?.dismiss();
    setBulkyTitle("");
    setBulkyMessage("");
    setBulkyImageURL("");
    setLoadingUpload(State.idle);
  }
  async function createOrUpdate() {
    if (!bulkyTitle) {
      setError("title");
      return;
    }
    if (!bulkyMessage) {
      setError("message");
      return;
    }
    if (!bulkyImageURL) {
      setError("image-url");
      return;
    }
    try {
      if (type == "create") {

        if (!imageFile) return;
        const fileID = await onUploadBulkyImage(imageFile);

        if(!fileID) return
        onSendBulkyItem(`${bulkyTitle}\n\n${bulkyMessage}`, fileID, () => {
          refScrollRoot.current?.scrollTo({
            top: 0,
          });
        });
      } else if (type == "update") {
        console.log("in update")

        // Update post
        console.log(fileID)
        if (!postID || !fileID) return;
        onUpdateBulkyItem(
          postID,
          `${bulkyTitle}\n\n${bulkyMessage}`,
          () => {
            refScrollRoot.current?.scrollTo({
              top: 0,
            });
          },
        );
      }

      let body: Parameters<typeof bulkyItemPut>[0] = {
        chain_uid: chain!.uid,
        user_uid: bulky?.user_uid || authUser!.uid,
        title: bulkyTitle,
        message: bulkyMessage,
        image_url: bulkyImageURL,
      };
      if (bulky) body.id = bulky.id;
      await bulkyItemPut(body);

      setError("");

      modal.current?.dismiss("", "confirm");
    } catch (err: any) {
      setError(err.status);
      toastError(present, err);
    }
  }

  function handleClickUpload() {
    if (isCapacitor) {
      setLoadingUpload(State.loading);
      handleNativeUpload()
        .then(() => {
          setLoadingUpload(State.success);
          setTimeout(() => {
            setLoadingUpload(State.idle);
          }, 2000);
        })
        .catch((err) => {
          toastError(present, err);
          setLoadingUpload(State.error);
        });
    } else {
      const el = document.getElementById(
        "cu-bulky-web-image-upload",
      ) as HTMLInputElement | null;
      el?.click();
    }
  }

  function handleWebUpload(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setImageFile(file);

    setLoadingUpload(State.loading);

    function getBase64(file: File) {
      return new Promise<string>((resolve, reject) => {
        let reader = new FileReader();
        reader.readAsDataURL(file);
        reader.onload = () =>
          resolve(
            (reader.result as string).replace("data:", "").replace(/^.+,/, ""),
          );
        reader.onerror = (error) => reject(error);
      });
    }

    (async () => {
      // https://pqina.nl/blog/convert-a-file-to-a-base64-string-with-javascript/#encoding-the-file-as-a-base-string
      const image64 = await getBase64(file);

      const res = await uploadImage(image64, 800);
      setBulkyImageURL(res.data.image);
    })()
      .then(() => {
        setLoadingUpload(State.success);
        setTimeout(() => {
          setLoadingUpload(State.idle);
        }, 2000);
      })
      .catch(() => {
        setLoadingUpload(State.error);
      });
  }

  async function handleNativeUpload() {
    const photo = await Camera.getPhoto({
      resultType: CameraResultType.Base64,
    });
    if (!photo.base64String) throw "Image not found";

    const res = await uploadImage(photo.base64String, 800);
    setBulkyImageURL(res.data.image);
  }

  return (
    <IonModal
      ref={modal}
      onIonModalWillPresent={modalInit}
      onIonModalDidDismiss={didDismiss}
    >
      <IonHeader>
        <IonToolbar>
          <IonButtons slot="start">
            <IonButton onClick={cancel}>{t("cancel")}</IonButton>
          </IonButtons>
          <IonTitle>
            {bulky ? t("updateBulkyItem") : t("createBulkyItem")}
          </IonTitle>
          <IonButtons slot="end">
            <IonButton
              onClick={createOrUpdate}
              color={!error ? "primary" : "danger"}
            >
              {t("save")}
            </IonButton>
          </IonButtons>
        </IonToolbar>
      </IonHeader>
      <IonContent fullscreen>
        <IonList>
          <IonItem color={error === "title" ? "danger" : undefined}>
            <IonInput
              type="text"
              autoCorrect="on"
              autoCapitalize="words"
              enterkeyhint="next"
              label={t("title")}
              labelPlacement="start"
              value={bulkyTitle}
              onIonInput={(e) => setBulkyTitle(e.detail.value + "")}
            ></IonInput>
          </IonItem>
          <IonItem
            lines="inset"
            color={error === "message" ? "danger" : undefined}
          >
            <IonTextarea
              className="ion-margin-bottom"
              label={t("message")}
              labelPlacement="start"
              spellCheck="true"
              autoGrow
              autoCapitalize="sentences"
              autoCorrect="on"
              enterkeyhint="next"
              value={bulkyMessage}
              onIonInput={(e) => setBulkyMessage(e.detail.value + "")}
            />
          </IonItem>
          <IonItem
            color={error === "image-url" ? "danger" : undefined}
            lines="none"
          >
            <div className="tw-w-full">
              <IonLabel className="tw-mt-2 tw-mb-0">{t("image")}</IonLabel>

              <div className="tw-text-center tw-w-full">
                {!(loadingUpload === State.loading) && bulkyImageURL ? (
                  <IonCard
                    onClick={handleClickUpload}
                    className={`tw-my-8 tw-mx-[50px] tw-border tw-border-solid ${
                      loadingUpload ? "tw-border-medium" : "tw-border-primary"
                    }`}
                  >
                    <IonImg
                      src={bulkyImageURL}
                      alt={t("loading")}
                      className="tw-max-w-full"
                    />
                  </IonCard>
                ) : (
                  <div className="tw-w-full tw-h-[300px] tw-flex tw-justify-center tw-items-center">
                    <IonCard
                      onClick={handleClickUpload}
                      className={`tw-border tw-border-solid tw-w-[200px] tw-h-[200px] tw-flex tw-justify-center tw-items-center ${
                        loadingUpload === State.loading
                          ? "tw-border-medium"
                          : "tw-border-primary"
                      }`}
                    >
                      <div className="tw-relative">
                        <IonIcon size="large" icon={imageOutline} />
                        {loadingUpload === State.loading ? (
                          <IonIcon
                            icon={hourglassOutline}
                            size="small"
                            color="primary"
                            className="tw-bg-primary-contrast tw-p-[2px] tw-rounded-full tw-absolute -tw-bottom-1 -tw-right-3"
                          />
                        ) : null}
                      </div>
                    </IonCard>
                  </div>
                )}
              </div>

              <IonButton
                onClick={handleClickUpload}
                size="default"
                className="tw-m-0 tw-mb-4"
                expand="block"
                color={
                  loadingUpload === State.idle
                    ? "primary"
                    : loadingUpload === State.loading
                    ? "light"
                    : loadingUpload === State.success
                    ? "success"
                    : "warning"
                }
              >
                <IonIcon
                  icon={
                    loadingUpload === State.loading
                      ? hourglassOutline
                      : loadingUpload === State.success
                      ? checkmarkOutline
                      : cloudUploadOutline
                  }
                  className="tw-mr-2"
                  size="default"
                />
                {loadingUpload === State.loading
                  ? t("loading")
                  : loadingUpload === State.error
                  ? "Error"
                  : loadingUpload === State.success
                  ? t("uploaded")
                  : t("upload")}
              </IonButton>
              {isCapacitor ? null : (
                <input
                  type="file"
                  id="cu-bulky-web-image-upload"
                  name="filename"
                  className="ion-hide"
                  onChange={handleWebUpload}
                />
              )}
            </div>
          </IonItem>
        </IonList>
      </IonContent>
    </IonModal>
  );
}
