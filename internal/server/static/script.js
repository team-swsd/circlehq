document.addEventListener('DOMContentLoaded', () => {
    const body = document.body;
    const tabButtons = document.querySelectorAll('.tab-button');
    const contentFrames = document.querySelectorAll('.content-frame');
    const reloadButton = document.getElementById('tab-reload-button');

    /**
     * タブ切り替え機能
     */
    tabButtons.forEach(button => {
        button.addEventListener('click', () => {
            activateTab(button);
        });
    });

    /**
     * 指定されたタブをアクティブにし、対応するiframeを表示する関数
     * @param {Element} tabToActivate - アクティブにするタブボタン要素
     */
    function activateTab(tabToActivate) {
        const targetId = tabToActivate.dataset.target;
        const targetFrame = document.getElementById(`${targetId}-frame`);

        // すべてのタブボタンから'active'クラスを削除
        tabButtons.forEach(btn => btn.classList.remove('active'));
        // クリックされたタブに'active'クラスを追加
        tabToActivate.classList.add('active');

        // すべてのiframeから'active'クラスを削除
        contentFrames.forEach(frame => frame.classList.remove('active'));
        // 対応するiframeに'active'クラスを追加
        targetFrame.classList.add('active');
    }

    /**
     * タブリロード機能
     */
    reloadButton.addEventListener('click', () => {
        // 現在表示されている（.activeクラスを持つ）iframeを取得
        const activeFrame = document.querySelector('.content-frame.active');

        // iframeが見つかった場合のみリロードを実行
        if (activeFrame) {
            // iframeのsrc属性を再設定することでリロードする
            // この方法はクロスオリジンのiframeでも機能します
            activeFrame.src = activeFrame.src;
        }
    });

    // 初期表示時に最初のタブをアクティブにする
    activateTab(tabButtons[0]);
});